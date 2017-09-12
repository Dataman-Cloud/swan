package strategy

import (
	magent "github.com/Dataman-Cloud/swan/mesos/agent"
)

type weightedAgent struct {
	agent  *magent.Agent
	weight float64
}

type weightedAgents []*weightedAgent

func (w weightedAgents) Len() int {
	return len(w)
}

func (w weightedAgents) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

func (w weightedAgents) Less(i, j int) bool {
	return w[i].weight < w[j].weight
}

func weight(agents []*magent.Agent) weightedAgents {
	weightedList := make([]*weightedAgent, 0)

	for _, agent := range agents {
		cpus, mem, disk, ports := agent.Resources()
		weighted := &weightedAgent{
			agent:  agent,
			weight: cpus + mem + disk + float64(len(ports)),
		}

		weightedList = append(weightedList, weighted)
	}

	return weightedList
}
