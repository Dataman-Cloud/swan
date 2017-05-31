package filter

import (
	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
)

type resourceFilter struct {
}

func NewResourceFilter() *resourceFilter {
	return &resourceFilter{}
}

func (f *resourceFilter) Filter(config *types.TaskConfig, agents []*mesos.Agent) []*mesos.Agent {
	candidates := make([]*mesos.Agent, 0)

	for _, agent := range agents {
		cpus, mem, disk, ports := agent.Resources()

		if cpus >= config.CPUs &&
			mem >= config.Mem &&
			disk >= config.Disk &&
			len(ports) >= len(config.PortMappings) {
			candidates = append(candidates, agent)
		}
	}

	return candidates
}
