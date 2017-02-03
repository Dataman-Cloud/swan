package janitor

import (
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	SWAN_UPSTREAM_LOADER_KEY = "SwanUpstreamLoader"
)

type TargetChangeEvent struct {
	Change string // "add" or "del"

	AppID    string
	TaskID   string
	TaskIP   string
	TaskPort uint32
	PortName string
}

type SwanUpstreamLoader struct {
	Upstreams     []*Upstream
	swanEventChan chan *TargetChangeEvent

	lock sync.Mutex
}

func SwanUpstreamLoaderInit() (*SwanUpstreamLoader, error) {
	swanUpstreamLoader := &SwanUpstreamLoader{
		Upstreams:     make([]*Upstream, 0),
		swanEventChan: make(chan *TargetChangeEvent, 1024),
	}

	go swanUpstreamLoader.Poll(context.Background()) // start polling at background, waiting targetChangeEvent arrive
	return swanUpstreamLoader, nil
}

func (swanUpstreamLoader *SwanUpstreamLoader) Poll(ctx context.Context) {
	log.Debug("SwanUpstreamLoader starts listening app event...")

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("SwanUpstreamLoader poll got error: %s", err)
			swanUpstreamLoader.Poll(context.Background()) // execute poll again
		}
	}()

	for {
		targetChangeEvent := <-swanUpstreamLoader.swanEventChan
		log.Debugf("SwanUpstreamLoader receive one app event:%s", targetChangeEvent)
		upstream := buildUpstream(targetChangeEvent)
		target := buildTarget(targetChangeEvent)

		switch strings.ToLower(targetChangeEvent.Change) {
		case "add":
			if swanUpstreamLoader.Contains(upstream) {
				u := swanUpstreamLoader.Get(upstream.AppID)
				if u == nil {
					panic("no upstream found")
				}

				// add target if not exists
				if !u.ContainsTarget(target.TaskID) {
					u.AddTarget(target)
				}
			} else {
				upstream.AddTarget(target)
				swanUpstreamLoader.Upstreams = append(swanUpstreamLoader.Upstreams, upstream)
			}
		case "del":
			if swanUpstreamLoader.Contains(upstream) {
				u := swanUpstreamLoader.Get(upstream.AppID)
				if u == nil {
					panic("no upstream found")
				}

				u.RemoveTarget(target)
				if len(u.Targets) == 0 { // remove upstream after last target was removed
					swanUpstreamLoader.RemoveUpstream(upstream)
				}
			}
		}
	}
}

func (swanUpstreamLoader *SwanUpstreamLoader) Contains(newUpstream *Upstream) bool {
	for _, u := range swanUpstreamLoader.Upstreams {
		if u.Equal(newUpstream) {
			return true
			break
		}
	}

	return false
}

func (swanUpstreamLoader *SwanUpstreamLoader) RemoveUpstream(upstream *Upstream) {
	index := -1
	for k, v := range swanUpstreamLoader.Upstreams {
		if v.Equal(upstream) {
			index = k
			break
		}
	}
	if index >= 0 {
		swanUpstreamLoader.lock.Lock()
		defer swanUpstreamLoader.lock.Unlock()
		swanUpstreamLoader.Upstreams = append(swanUpstreamLoader.Upstreams[:index], swanUpstreamLoader.Upstreams[index+1:]...)
	}
}

func (swanUpstreamLoader *SwanUpstreamLoader) List() []*Upstream {
	swanUpstreamLoader.lock.Lock()
	defer swanUpstreamLoader.lock.Unlock()
	return swanUpstreamLoader.Upstreams
}

func (swanUpstreamLoader *SwanUpstreamLoader) SwanEventChan() chan<- *TargetChangeEvent {
	return swanUpstreamLoader.swanEventChan
}

func (swanUpstreamLoader *SwanUpstreamLoader) Get(appID string) *Upstream {
	for _, u := range swanUpstreamLoader.Upstreams {
		if u.AppID == appID {
			return u
			break
		}
	}
	return nil
}

func buildTarget(targetChangeEvent *TargetChangeEvent) *Target {
	return &Target{
		AppID:    targetChangeEvent.AppID,
		TaskID:   targetChangeEvent.TaskID,
		TaskIP:   targetChangeEvent.TaskIP,
		TaskPort: targetChangeEvent.TaskPort,
		PortName: targetChangeEvent.PortName,
	}
}

func buildUpstream(targetChangeEvent *TargetChangeEvent) *Upstream {
	up := NewUpstream()
	up.Targets = make([]*Target, 0)
	up.lock = sync.Mutex{}
	up.AppID = targetChangeEvent.AppID

	return up
}
