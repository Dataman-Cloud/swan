package filter

import (
	magent "github.com/Dataman-Cloud/swan/mesos/agent"
	"github.com/Dataman-Cloud/swan/types"
)

type Filter interface {
	Filter(opts *FilterOptions, agents []*magent.Agent) ([]*magent.Agent, error)
}

type FilterOptions struct {
	// resources requirements
	ResRequired types.ResourcesRequired
	Replicas    int

	// constraints
	Constraints []*types.Constraint
}

// the returned agents contains at least one proper agent
func ApplyFilters(filters []Filter, opts *FilterOptions, agents []*magent.Agent) ([]*magent.Agent, error) {
	accepted := agents

	var err error
	for _, filter := range filters {
		accepted, err = filter.Filter(opts, accepted)
		if err != nil {
			return nil, err
		}
	}

	return accepted, nil
}
