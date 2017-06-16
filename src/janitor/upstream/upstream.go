package upstream

import (
	"fmt"
	"sync"
)

var mgr *UpsManager

func init() {
	mgr = &UpsManager{
		Upstreams: make([]*Upstream, 0, 0),
	}
}

type UpsManager struct {
	Upstreams []*Upstream `json:"upstreams"`
	sync.RWMutex
}

type Upstream struct {
	AppID     string    `json:"app_id"` // uniq id of upstream
	AppAlias  string    `json:"app_alias"`
	AppListen string    `json:"app_listen"`
	Targets   []*Target `json:"targets"`

	sessions *Sessions
	balancer Balancer
}

func (u *Upstream) search(taskID string) (int, *Target) {
	for i, v := range u.Targets {
		if v.TaskID == taskID {
			return i, v
		}
	}
	return -1, nil
}

func newUpstream(first *Target) *Upstream {
	return &Upstream{
		AppID:     first.AppID,
		AppAlias:  first.AppAlias,
		AppListen: first.AppListen,
		Targets:   []*Target{first},
		balancer:  &weightBalancer{}, // default balancer
		sessions:  newSessions(),     // sessions store
	}
}

func AllUpstreams() []*Upstream {
	mgr.RLock()
	defer mgr.RUnlock()
	return mgr.Upstreams
}

func AllSessions() map[string]*Sessions {
	mgr.RLock()
	defer mgr.RUnlock()

	ret := make(map[string]*Sessions)
	for _, u := range mgr.Upstreams {
		ret[u.AppID] = u.sessions
	}
	return ret
}

func UpsertTarget(target *Target) (onFirst bool, err error) {
	mgr.Lock()
	defer mgr.Unlock()

	var (
		appID     = target.AppID
		appAlias  = target.AppAlias
		appListen = target.AppListen
		taskID    = target.TaskID
	)

	_, u := getUpstreamByID(appID)
	// add new upstream
	if u == nil {
		onFirst = true

		if i, _ := getUpstreamByAlias(appAlias); i >= 0 {
			err = fmt.Errorf("alias address [%s] conflict", appAlias)
			return
		}
		if i, _ := getUpstreamByListen(appListen); i >= 0 {
			err = fmt.Errorf("listen address [%s] conflict", appListen)
			return
		}

		mgr.Upstreams = append(mgr.Upstreams, newUpstream(target))
		return
	}

	_, t := u.search(taskID)

	// add new target
	if t == nil {
		u.Targets = append(u.Targets, target)
		return
	}

	// update target
	t.VersionID = target.VersionID
	t.AppVersion = target.AppVersion
	t.TaskIP = target.TaskIP
	t.TaskPort = target.TaskPort
	t.Scheme = ""
	t.Weight = target.Weight

	return
}

func GetTarget(appID, taskID string) *Target {
	mgr.RLock()
	defer mgr.RUnlock()

	_, u := getUpstreamByID(appID)
	if u == nil {
		return nil
	}

	_, t := u.search(taskID)
	return t
}

func RemoveTarget(target *Target) (onLast bool) {
	mgr.Lock()
	defer mgr.Unlock()

	var (
		appID  = target.AppID
		taskID = target.TaskID
	)

	idxu, u := getUpstreamByID(appID)
	if u == nil {
		return
	}

	idxt, t := u.search(taskID)
	if t == nil {
		return
	}

	// remove target & session
	u.Targets = append(u.Targets[:idxt], u.Targets[idxt+1:]...)
	u.sessions.remove(taskID)

	// remove empty upstream & stop sessions gc
	if len(u.Targets) == 0 {
		onLast = true
		u.sessions.stop()
		mgr.Upstreams = append(mgr.Upstreams[:idxu], mgr.Upstreams[idxu+1:]...)
	}

	return
}

// similar as lookup, but by app alias
func LookupAlias(remoteIP, appAlias string) *Target {
	mgr.RLock()
	_, u := getUpstreamByAlias(appAlias)
	mgr.RUnlock()

	if u == nil {
		return nil
	}

	appID := u.AppID
	return Lookup(remoteIP, appID, "")
}

// similar as lookup, but by app listen
func LookupListen(remoteIP, appListen string) *Target {
	mgr.RLock()
	_, u := getUpstreamByListen(appListen)
	mgr.RUnlock()

	if u == nil {
		return nil
	}

	appID := u.AppID
	return Lookup(remoteIP, appID, "")
}

// lookup select a suitable backend according by sessions & balancer
func Lookup(remoteIP, appID, taskID string) *Target {
	var (
		u *Upstream
		t *Target
	)

	if _, u = getUpstreamByID(appID); u == nil {
		return nil
	}

	defer func() {
		if t != nil {
			u.sessions.update(remoteIP, t)
		}
	}()

	// obtain specified task backend
	if taskID != "" {
		t = GetTarget(appID, taskID)
		return t
	}

	// obtain session by remoteIP
	if t = u.sessions.get(remoteIP); t != nil {
		return t
	}

	// use balancer to obtain a new backend
	t = nextTarget(appID)
	return t
}

func nextTarget(appID string) *Target {
	mgr.RLock()
	defer mgr.RUnlock()

	_, u := getUpstreamByID(appID)
	if u == nil {
		return nil
	}

	return u.balancer.Next(u.Targets)
}

// note: must be called under protection of mutext lock
func getUpstreamByID(appID string) (int, *Upstream) {
	for i, v := range mgr.Upstreams {
		if v.AppID == appID {
			return i, v
		}
	}
	return -1, nil
}

// note: must be called under protection of mutext lock
func getUpstreamByAlias(alias string) (int, *Upstream) {
	if alias == "" {
		return -1, nil
	}
	for i, v := range mgr.Upstreams {
		if v.AppAlias == alias {
			return i, v
		}
	}
	return -1, nil
}

// note: must be called under protection of mutext lock
func getUpstreamByListen(listen string) (int, *Upstream) {
	if listen == "" {
		return -1, nil
	}
	for i, v := range mgr.Upstreams {
		if v.AppListen == listen {
			return i, v
		}
	}
	return -1, nil
}
