package janitor

import (
	"net/url"
	"sync"
)

type Upstream struct {
	AppID string

	Targets     []*Target
	LoadBalance *RoundRobinLoadBalancer

	lock sync.Mutex
}

func (u *Upstream) GetTarget(taskID string) *Target {
	for _, t := range u.Targets {
		if t.TaskID == taskID {
			return t
			break
		}
	}
	return nil
}

func NewUpstream() *Upstream {
	lb := NewRoundRobinLoadBalancer()
	lb.Seed()
	return &Upstream{
		LoadBalance: lb,
	}
}

func (u *Upstream) Equal(o *Upstream) bool {
	return u.AppID == o.AppID
}

func (u *Upstream) ContainsTarget(taskID string) bool {
	return u.GetTarget(taskID) != nil
}

func (u *Upstream) AddTarget(target *Target) {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.Targets = append(u.Targets, target)
}

func (u *Upstream) RemoveTarget(target *Target) {
	index := -1
	for k, v := range u.Targets {
		if v.Equal(target) {
			index = k
			break
		}
	}
	if index >= 0 {
		u.lock.Lock()
		defer u.lock.Unlock()
		u.Targets = append(u.Targets[:index], u.Targets[index+1:]...)
	}
}

func (u *Upstream) NextTargetEntry() *url.URL {
	u.lock.Lock()
	defer u.lock.Unlock()

	rr := u.LoadBalance
	current := u.Targets[rr.NextIndex]
	rr.NextIndex = (rr.NextIndex + 1) % len(u.Targets)

	return current.Entry()
}
