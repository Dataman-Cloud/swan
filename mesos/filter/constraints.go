package filter

import (
	"errors"

	magent "github.com/Dataman-Cloud/swan/mesos/agent"
)

var (
	errNoSatisfiedAgent = errors.New("no satisfied agent")
)

type constraintsFilter struct{}

func NewConstraintsFilter() *constraintsFilter {
	return &constraintsFilter{}
}

func (f *constraintsFilter) Filter(opts *FilterOptions, agents []*magent.Agent) ([]*magent.Agent, error) {
	var (
		constraints = opts.Constraints
		candidates  = make([]*magent.Agent, 0)
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
