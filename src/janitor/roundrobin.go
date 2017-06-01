package janitor

import (
	"sync"
)

type LoadBalancer interface {
	Seed([]*Target) *Target
}

type RoundRobinLoadBalancer struct {
	//Upstream  *upstream.Upstream
	nextIndex uint
	mu        sync.Mutex
}

func NewRoundRobinLoadBalancer() *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{
		nextIndex: 0,
	}
}

func (rr *RoundRobinLoadBalancer) Seed(targets []*Target) *Target {
	target := targets[rr.nextIndex]
	rr.mu.Lock()
	rr.nextIndex = (rr.nextIndex + 1) % uint(len(targets))
	rr.mu.Unlock()

	return target
}
