package filter

import (
	"errors"

	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
)

var (
	errResourceNotEnough = errors.New("resource not enough")
)

type resourceFilter struct {
}

func NewResourceFilter() *resourceFilter {
	return &resourceFilter{}
}

// multiplicate with replicas to calculate total resource requirments
func (f *resourceFilter) Filter(config *types.TaskConfig, replicas int, agents []*mesos.Agent) ([]*mesos.Agent, error) {
	candidates := make([]*mesos.Agent, 0)

	for _, agent := range agents {
		var (
			cpus, mem, disk, ports = agent.Resources()               // avaliable agent resources
			cpusReq                = config.CPUs * float64(replicas) // total resources requirements ...
			memReq                 = config.Mem * float64(replicas)
			diskReq                = config.Disk * float64(replicas)
			portsReq               = len(config.PortMappings) * replicas // FIXME LATER
		)
		if cpus >= cpusReq && mem >= memReq && disk >= diskReq && len(ports) >= portsReq {
			candidates = append(candidates, agent)
		}
	}

	if len(candidates) == 0 {
		return nil, errResourceNotEnough
	}
	return candidates, nil
}
