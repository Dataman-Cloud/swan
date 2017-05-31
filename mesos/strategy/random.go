package strategy

import (
	"math/rand"
	"time"

	"github.com/Dataman-Cloud/swan/mesos"
)

type randomStrategy struct {
	r *rand.Rand
}

func NewRandomStrategy() *randomStrategy {
	return &randomStrategy{
		r: rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
	}
}

func (m *randomStrategy) RankAndSort(agents []*mesos.Agent) []*mesos.Agent {
	for i := 0; i < len(agents); i++ {
		j := m.r.Intn(i + 1)
		agents[i], agents[j] = agents[j], agents[i]
	}

	return agents
}
