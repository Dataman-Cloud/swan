package filter

import (
	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/store"
	"github.com/Dataman-Cloud/swan/types"
)

type constraintsFilter struct {
	db store.Store
}

func NewConstraintsFilter(db store.Store) *constraintsFilter {
	return &constraintsFilter{
		db: db,
	}
}

func (f *constraintsFilter) Filter(config *types.TaskConfig, agents []*mesos.Agent) []*mesos.Agent {
	constraints := config.Constraints

	candidates := make([]*mesos.Agent, 0)

	for _, agent := range agents {
		match := true
		for _, constraint := range constraints {
			attrs := merge(
				f.getAttrs(config.Cluster, agent.IP()),
				agent.Attributes(),
			)
			if constraint.Match(attrs) {
				continue
			}
			match = false
			break
		}

		if match {
			candidates = append(candidates, agent)
		}
	}

	return candidates
}

func (f *constraintsFilter) getAttrs(clusterId, agentIp string) map[string]string {
	s, err := f.db.GetNode(clusterId, agentIp)
	if err != nil {
		return nil
	}

	return s.Attrs
}

func merge(a map[string]string, b map[string]string) map[string]string {
	c := make(map[string]string)

	for k, v := range a {
		c[k] = v
	}

	for k, v := range b {
		if _, ok := c[k]; !ok {
			c[k] = v
		}
	}

	return c
}
