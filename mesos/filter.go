package mesos

import (
	"github.com/Dataman-Cloud/swan/types"
)

type Filter interface {
	Filter(config *types.TaskConfig, replicas int, agents []*Agent) ([]*Agent, error)
}

// the returned agents contains at least one proper agent
func ApplyFilters(filters []Filter, config *types.TaskConfig, replicas int, agents []*Agent) ([]*Agent, error) {
	accepted := agents

	var err error
	for _, filter := range filters {
		accepted, err = filter.Filter(config, replicas, accepted)
		if err != nil {
			return nil, err
		}

	}

	return accepted, nil
}
