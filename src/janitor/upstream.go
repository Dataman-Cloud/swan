package janitor

import (
	"fmt"
	"net/url"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type Upstreams struct {
	Upstreams []*Upstream `json:"upstreams"`
	sync.RWMutex
}

type Upstream struct {
	AppID    string    `json:"app_id"` // uniq id of upstream
	Targets  []*Target `json:"targets"`
	sessions *Sessions
	balancer Balancer
}

func newUpstream(appID string) *Upstream {
	return &Upstream{
		AppID:    appID,
		Targets:  make([]*Target, 0, 0),
		balancer: &WeightBalancer{}, // default balancer
		sessions: newSessions(),     // sessions store
	}
}

func (us *Upstreams) allUps() []*Upstream {
	us.RLock()
	defer us.RUnlock()
	return us.Upstreams
}

func (us *Upstreams) allSess() map[string]*Sessions {
	us.RLock()
	defer us.RUnlock()

	ret := make(map[string]*Sessions)
	for _, u := range us.Upstreams {
		ret[u.AppID] = u.sessions
	}
	return ret
}

func (us *Upstreams) addTarget(appID string, target *Target) {
	us.Lock()
	defer us.Unlock()

	if target == nil {
		return
	}

	_, u := us.getUpstream(appID)
	if u == nil { // add new upstream
		u = newUpstream(appID)
		u.Targets = append(u.Targets, target)
		us.Upstreams = append(us.Upstreams, u)
		return
	}

	_, t := u.getTarget(target.TaskID)
	if t != nil {
		log.Warnf("already exists the target %v, ignore.", *t)
		return
	}

	u.Targets = append(u.Targets, target)
}

func (us *Upstreams) getTarget(appID, taskID string) *Target {
	us.RLock()
	defer us.RUnlock()

	_, u := us.getUpstream(appID)
	if u == nil {
		return nil
	}

	_, t := u.getTarget(taskID)
	return t
}

func (us *Upstreams) removeTarget(appID, taskID string) {
	us.Lock()
	defer us.Unlock()

	idxu, u := us.getUpstream(appID)
	if u == nil {
		log.Warnln("no such upstream", appID)
		return
	}

	idxt, t := u.getTarget(taskID)
	if t == nil {
		log.Warnf("no such target", taskID)
		return
	}

	// remove target
	u.Targets = append(u.Targets[:idxt], u.Targets[idxt+1:]...)
	u.sessions.remove(taskID)

	// remove empty upstream
	if len(u.Targets) == 0 {
		u.sessions.stop() // stop sessions gc & clean up
		us.Upstreams = append(us.Upstreams[:idxu], us.Upstreams[idxu+1:]...)
	}
}

func (us *Upstreams) updateTarget(appID string, new *Target) {
	us.Lock()
	defer us.Unlock()

	_, u := us.getUpstream(appID)
	if u == nil {
		log.Warnln("no such upstream", appID)
		return
	}

	_, t := u.getTarget(new.TaskID)
	if t == nil {
		log.Warnf("no such target", new.TaskID)
		return
	}

	t.Weight = new.Weight // NOTE only update weight currently
}

// lookup select a suitable backend according by sessions & balancer
func (us *Upstreams) lookup(remoteIP, appID, taskID string) *Target {
	var (
		u *Upstream
		t *Target
	)

	if _, u = us.getUpstream(appID); u == nil {
		return nil
	}

	defer func() {
		if t != nil {
			u.sessions.update(remoteIP, t)
		}
	}()

	// obtain session
	if t = u.sessions.get(remoteIP); t != nil {
		return t
	}

	// obtain specified task backend
	if taskID != "" {
		t = us.getTarget(appID, taskID)
		return t
	}

	// use balancer to obtain a new backend
	t = us.nextTarget(appID)
	return t
}

func (us *Upstreams) nextTarget(appID string) *Target {
	us.RLock()
	defer us.RUnlock()

	_, u := us.getUpstream(appID)
	if u == nil {
		return nil
	}

	return u.balancer.Next(u.Targets)
}

// note: must be called under protection of mutext lock
func (us *Upstreams) getUpstream(appID string) (int, *Upstream) {
	for i, v := range us.Upstreams {
		if v.AppID == appID {
			return i, v
		}
	}
	return -1, nil
}

// note: must be called under protection of mutext lock
func (u *Upstream) getTarget(taskID string) (int, *Target) {
	for i, v := range u.Targets {
		if v.TaskID == taskID {
			return i, v
		}
	}
	return -1, nil
}

// Target
type Target struct {
	AppID      string  `json:"app_id"`
	VersionID  string  `json:"version_id"`
	AppVersion string  `json:"app_version"`
	TaskID     string  `json:"task_id"`
	TaskIP     string  `json:"task_ip"`
	TaskPort   uint32  `json:"task_port"`
	PortName   string  `json:"port_name"`
	Weight     float64 `json:"weihgt"`
}

func (t Target) url() *url.URL {
	s := fmt.Sprintf("http://%s:%d", t.TaskIP, t.TaskPort)
	u, err := url.Parse(s)
	if err != nil {
		log.Errorf("invalid task url entry %s - [%v]", s, err)
		return nil
	}

	return u
}

// TargetChangeEvent
type TargetChangeEvent struct {
	Change string // add/del/update
	Target
}

func (ev TargetChangeEvent) String() string {
	return fmt.Sprintf("{%s: app:%s task:%s ip:%s:%d weight:%f}",
		ev.Change, ev.AppID, ev.TaskID, ev.TaskIP, ev.TaskPort, ev.Weight)
}
