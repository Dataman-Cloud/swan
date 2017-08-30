package zk

import (
	"fmt"
	"path"

	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
)

func (zk *ZKStore) CreateMesosAgent(agent *types.MesosAgent) error {
	p := path.Join(keyAgent, agent.IP)

	bs, err := encode(agent)
	if err != nil {
		return err
	}

	return zk.create(p, bs)
}

func (zk *ZKStore) GetMesosAgent(agentId string) (*types.MesosAgent, error) {
	p := path.Join(keyAgent, agentId)

	data, _, err := zk.get(p)
	if err != nil {
		if err == errNotExists {
			return nil, fmt.Errorf("agent %s not exists", agentId)
		}

		return nil, err
	}

	var agent types.MesosAgent
	if err := decode(data, &agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

func (zk *ZKStore) UpdateMesosAgent(agent *types.MesosAgent) error {
	bs, err := encode(agent)
	if err != nil {
		return err
	}

	p := path.Join(keyAgent, agent.IP)

	return zk.set(p, bs)
}

func (zk *ZKStore) ListMesosAgents() ([]*types.MesosAgent, error) {
	nodes, err := zk.list(keyAgent)
	if err != nil {
		log.Errorln("zk ListMesosAgents error:", err)
		return nil, err
	}

	agents := make([]*types.MesosAgent, 0)
	for _, node := range nodes {
		agent, err := zk.GetMesosAgent(node)
		if err != nil {
			log.Errorln("zk GetMesosAgent error: %v", err)
			continue
		}

		agents = append(agents, agent)
	}

	return agents, nil
}
