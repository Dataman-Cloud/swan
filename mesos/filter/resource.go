package filter

import (
	"errors"

	magent "github.com/Dataman-Cloud/swan/mesos/agent"
)

var (
	errResourceNotEnough = errors.New("resource not enough")
)

type resourceFilter struct{}

func NewResourceFilter() *resourceFilter {
	return &resourceFilter{}
}

func (f *resourceFilter) Filter(opts *FilterOptions, agents []*magent.Agent) ([]*magent.Agent, error) {
	var (
		candidates = make([]*magent.Agent, 0)
		replicas   = opts.Replicas

		// multiplicate with replicas to calculate total resource requirments
		cpuReq   = opts.ResRequired.CPUs * float64(replicas)
		memReq   = opts.ResRequired.Mem * float64(replicas)
		diskReq  = opts.ResRequired.Disk * float64(replicas)
		portsReq = opts.ResRequired.NumPort * replicas
	)

	for _, agent := range agents {
		var (
			cpus, mem, disk, ports = agent.Resources() // avaliable agent resources
		)
		if cpus >= cpuReq && mem >= memReq && disk >= diskReq && len(ports) >= portsReq {
			candidates = append(candidates, agent)
		}
	}

	if len(candidates) == 0 {
		return nil, errResourceNotEnough
	}
	return candidates, nil
}
