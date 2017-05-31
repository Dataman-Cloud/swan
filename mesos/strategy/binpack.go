package strategy

import (
	"sort"

	"github.com/Dataman-Cloud/swan/mesos"
)

type binpackStrategy struct{}

func NewBinPackStrategy() *binpackStrategy {
	return &binpackStrategy{}
}

func (b *binpackStrategy) RankAndSort(agents []*mesos.Agent) []*mesos.Agent {
	weightedList := weight(agents)

	sort.Sort(weightedList)

	candidates := make([]*mesos.Agent, 0)

	for _, weighted := range weightedList {
		candidates = append(candidates, weighted.agent)
	}

	return candidates
}
