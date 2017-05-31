package mesos

import (
	//"github.com/Dataman-Cloud/swan/mesos/filter"
	"github.com/Dataman-Cloud/swan/types"
)

type Filter interface {
	Filter(config *types.TaskConfig, agents []*Agent) []*Agent
}

//func NewFilter() []Filter {
//	filters := []Filter{
//		filter.NewResourceFilter(),
//	}
//
//	return filters
//}

func ApplyFilters(filters []Filter, config *types.TaskConfig, agents []*Agent) []*Agent {
	accepted := agents

	for _, filter := range filters {
		accepted = filter.Filter(config, accepted)
	}

	return accepted
}
