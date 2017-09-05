package filter

import (
	"errors"

	magent "github.com/Dataman-Cloud/swan/mesos/agent"
	"github.com/Dataman-Cloud/swan/store"
	"github.com/Dataman-Cloud/swan/types"
)

var (
	errNoSatisfiedAgent = errors.New("no satisfied agent")
)

type constraintsFilter struct {
	db store.Store
}

func NewConstraintsFilter(db store.Store) *constraintsFilter {
	return &constraintsFilter{
		db: db,
	}
}

func (f *constraintsFilter) Filter(config *types.TaskConfig, replicas int, agents []*magent.Agent) ([]*magent.Agent, error) {
	var (
		constraints = config.Constraints
		candidates  = make([]*magent.Agent, 0)
	)

	for _, agent := range agents {
		match := true
		for _, constraint := range constraints {
			attrs := f.getAttrs(agent.IP())
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

	if len(candidates) == 0 {
		return nil, errNoSatisfiedAgent
	}
	return candidates, nil
}

func (f *constraintsFilter) getAttrs(agentIp string) map[string]string {
	s, err := f.db.GetMesosAgent(agentIp)
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
