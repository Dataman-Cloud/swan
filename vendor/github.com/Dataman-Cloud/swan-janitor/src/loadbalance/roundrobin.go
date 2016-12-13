package loadbalance

import (
	"sync"
)

type LoadBalancer interface {
	//Next() *upstream.Target
	//Seed(upstream *upstream.Upstream)
}

type RoundRobinLoadBalancer struct {
	//Upstream  *upstream.Upstream
	NextIndex int
	SeedLock  sync.Mutex
}

func NewRoundRobinLoadBalancer() *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{}
}

func (rr *RoundRobinLoadBalancer) Seed() {
	rr.SeedLock.Lock()
	defer rr.SeedLock.Unlock()
	rr.NextIndex = 0
}
