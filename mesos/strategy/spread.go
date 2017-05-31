package strategy

import (
	"sort"

	"github.com/Dataman-Cloud/swan/mesos"
)

type spreadStrategy struct{}

func NewSpreadStrategy() *spreadStrategy {
	return &spreadStrategy{}
}

func (s *spreadStrategy) RankAndSort(agents []*mesos.Agent) []*mesos.Agent {
	weightedList := weight(agents)

	sort.Sort(sort.Reverse(weightedList))

	candidates := make([]*mesos.Agent, 0)

	for _, weighted := range weightedList {
		candidates = append(candidates, weighted.agent)
	}

	return candidates
}
