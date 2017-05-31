package filter

import (
	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
)

type constraintsFilter struct{}

func NewConstraintsFilter() *constraintsFilter {
	return &constraintsFilter{}
}

func (f *constraintsFilter) Filter(config *types.TaskConfig, agents []*mesos.Agent) []*mesos.Agent {
	constraints := config.Constraints

	candidates := make([]*mesos.Agent, 0)

	for _, agent := range agents {
		match := true
		for _, constraint := range constraints {
			if constraint.Match(agent.Attributes()) {
				continue
			}
			match = false
			break
		}

		if match {
			candidates = append(candidates, agent)
		}
	}

	return candidates
}
