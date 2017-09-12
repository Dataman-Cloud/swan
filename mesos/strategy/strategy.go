package strategy

import (
	magent "github.com/Dataman-Cloud/swan/mesos/agent"
)

type Strategy interface {
	RankAndSort(agents []*magent.Agent) []*magent.Agent
}
