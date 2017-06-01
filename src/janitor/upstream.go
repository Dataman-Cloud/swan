package janitor

import "sync"

type Upstream struct {
	AppID string `json:"appID"`

	Targets      []*Target `json:"targets"`
	loadBalancer LoadBalancer

	sync.RWMutex // protect Targets
}

func NewUpstream() *Upstream {
	return &Upstream{
		loadBalancer: NewWeightLoadBalancer(),
	}
}

func (u *Upstream) AddTarget(target *Target) {
	u.Lock()
	u.Targets = append(u.Targets, target)
	u.Unlock()
}

func (u *Upstream) UpdateTargetWeight(taskID string, newWeight float64) {
	u.Lock()
	defer u.Unlock()

	var target *Target
	for _, t := range u.Targets {
		if formatID(t.TaskID) == formatID(taskID) {
			target = t
		}
	}

	if target != nil {
		target.Weight = newWeight
	}
}

func (u *Upstream) RemoveTarget(target *Target) {
	u.Lock()
	defer u.Unlock()

	for i, v := range u.Targets {
		if v.Equal(target) {
			u.Targets = append(u.Targets[:i], u.Targets[i+1:]...)
			return
		}
	}
}

func (u *Upstream) NextTargetEntry() *Target {
	return u.loadBalancer.Seed(u.Targets)
}

func (u *Upstream) GetTarget(taskID string) *Target {
	u.RLock()
	defer u.RUnlock()

	for _, t := range u.Targets {
		if formatID(t.TaskID) == formatID(taskID) {
			return t
		}
	}

	return nil
}
