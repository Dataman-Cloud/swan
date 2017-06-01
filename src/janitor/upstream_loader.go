package janitor

import (
	"errors"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type TargetChangeEvent struct {
	Change string // "add" or "del" or "update"
	Target
}

type UpstreamLoader struct {
	Upstreams []*Upstream `json:"upstreams"`

	eventChan    chan *TargetChangeEvent
	sync.RWMutex // protect Upstreams
}

func NewUpstreamLoader(eventChan chan *TargetChangeEvent) *UpstreamLoader {
	return &UpstreamLoader{
		Upstreams: make([]*Upstream, 0),
		eventChan: eventChan,
	}
}

func (loader *UpstreamLoader) Start() error {
	log.Println("UpstreamLoader starts listening app event...")

	for ev := range loader.eventChan {
		log.Debugf("upstreamLoader receive app event: %+v", ev)

		upstream := ev.NewUpstream()
		target := &ev.Target

		switch strings.ToLower(ev.Change) {
		case "add":
			if u := loader.Get(upstream.AppID); u != nil {
				if u.GetTarget(target.TaskID) == nil {
					u.AddTarget(target)
				}
			} else {
				upstream.AddTarget(target)
				loader.Lock()
				loader.Upstreams = append(loader.Upstreams, upstream)
				loader.Unlock()
			}

		case "del":
			if u := loader.Get(upstream.AppID); u != nil {
				u.RemoveTarget(target)
				if len(u.Targets) == 0 { // remove upstream after last target was removed
					loader.RemoveUpstream(upstream)
				}
			}

			// update, weight only for the time present
		case "change":
			if u := loader.Get(upstream.AppID); u != nil {
				t := u.GetTarget(target.TaskID)
				if t == nil {
					log.Errorf("failed to find target %s from upstream", target.TaskID)
					break
				}
				u.UpdateTargetWeight(target.TaskID, target.Weight)
			}

		default:
			log.Warnln("unrecognized event change type", ev.Change)
		}
	}

	return errors.New("UpstreamLoader.eventChan closed")
}

func (loader *UpstreamLoader) RemoveUpstream(upstream *Upstream) {
	loader.Lock()
	defer loader.Unlock()

	for i, v := range loader.Upstreams {
		if v.AppID == upstream.AppID {
			loader.Upstreams = append(loader.Upstreams[:i], loader.Upstreams[i+1:]...)
			return
		}
	}
}

func (loader *UpstreamLoader) List() []*Upstream {
	loader.RLock()
	defer loader.RUnlock()
	return loader.Upstreams
}

func (loader *UpstreamLoader) Get(appID string) *Upstream {
	loader.RLock()
	defer loader.RUnlock()

	for _, u := range loader.Upstreams {
		if u.AppID == appID {
			return u
		}
	}

	return nil
}

func (ev *TargetChangeEvent) NewUpstream() *Upstream {
	return &Upstream{
		AppID:        ev.AppID,
		Targets:      make([]*Target, 0),
		loadBalancer: NewWeightLoadBalancer(),
	}
}
