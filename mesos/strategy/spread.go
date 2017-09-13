package strategy

import (
	"sort"

	magent "github.com/Dataman-Cloud/swan/mesos/agent"
)

type spreadStrategy struct{}

func NewSpreadStrategy() *spreadStrategy {
	return &spreadStrategy{}
}

func (s *spreadStrategy) RankAndSort(agents []*magent.Agent) []*magent.Agent {
	weightedList := weight(agents)

	sort.Sort(sort.Reverse(weightedList))

	candidates := make([]*magent.Agent, 0)

	for _, weighted := range weightedList {
		candidates = append(candidates, weighted.agent)
	}

	return candidates
}
