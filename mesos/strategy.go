package mesos

import ()

type Strategy interface {
	RankAndSort(agents []*Agent) []*Agent
}
