package strategy

import (
	"sort"

	magent "github.com/Dataman-Cloud/swan/mesos/agent"
)

type binpackStrategy struct{}

func NewBinPackStrategy() *binpackStrategy {
	return &binpackStrategy{}
}

func (b *binpackStrategy) RankAndSort(agents []*magent.Agent) []*magent.Agent {
	weightedList := weight(agents)

	sort.Sort(weightedList)

	candidates := make([]*magent.Agent, 0)

	for _, weighted := range weightedList {
		candidates = append(candidates, weighted.agent)
	}

	return candidates
}
