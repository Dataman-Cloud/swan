package janitor

import (
	"sync"
)

type Upstream struct {
	AppID string `json:"appID"`

	Targets      []*Target `json:"targets"`
	loadBalancer LoadBalancer

	mu sync.RWMutex
}

func NewUpstream() *Upstream {
	return &Upstream{
		loadBalancer: NewWeightLoadBalancer(),
	}
}

func (u *Upstream) Equal(o *Upstream) bool {
	return u.AppID == o.AppID
}

func (u *Upstream) ContainsTarget(taskID string) bool {
	return u.GetTarget(taskID) != nil
}

func (u *Upstream) AddTarget(target *Target) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.Targets = append(u.Targets, target)
}

func (u *Upstream) UpdateTargetWeight(taskID string, newWeight float64) {
	target := u.GetTarget(taskID)
	if target != nil {
		target.Weight = newWeight
	}
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
		u.mu.Lock()
		defer u.mu.Unlock()

		u.Targets = append(u.Targets[:index], u.Targets[index+1:]...)
	}
}

func (u *Upstream) NextTargetEntry() *Target {
	return u.loadBalancer.Seed(u.Targets)
}

func (u *Upstream) GetTarget(taskID string) *Target {
	for _, t := range u.Targets {
		if t.TaskID == taskID {
			return t
		}
	}

	return nil
}
