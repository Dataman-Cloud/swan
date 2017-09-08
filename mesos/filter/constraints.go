package filter

import (
	"errors"

	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
)

var (
	errNoSatisfiedAgent = errors.New("no satisfied agent")
)

type constraintsFilter struct{}

func NewConstraintsFilter() *constraintsFilter {
	return &constraintsFilter{}
}

func (f *constraintsFilter) Filter(config *types.TaskConfig, replicas int, agents []*mesos.Agent) ([]*mesos.Agent, error) {
	var (
		constraints = config.Constraints
		candidates  = make([]*mesos.Agent, 0)
	)

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

	if len(candidates) == 0 {
		return nil, errNoSatisfiedAgent
	}
	return candidates, nil
}
