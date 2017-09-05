package etcd

import (
	"fmt"
	"path"

	"github.com/Dataman-Cloud/swan/types"
	log "github.com/Sirupsen/logrus"
)

func (es *EtcdStore) CreateMesosAgent(agent *types.MesosAgent) error {
	p := path.Join(keyAgent, agent.IP)

	bs, err := encode(agent)
	if err != nil {
		return err
	}

	return es.create(p, bs)
}

func (es *EtcdStore) GetMesosAgent(agentIp string) (*types.MesosAgent, error) {
	p := path.Join(keyAgent, agentIp)

	data, err := es.get(p)
	if err != nil {
		if es.IsErrNotFound(err) {
			return nil, fmt.Errorf("agent %s not found", agentIp)
		}

		return nil, err
	}

	var agent types.MesosAgent
	if err := decode(data, &agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

func (es *EtcdStore) UpdateMesosAgent(agent *types.MesosAgent) error {
	bs, err := encode(agent)
	if err != nil {
		return err
	}

	p := path.Join(keyAgent, agent.IP, "value")

	return es.update(p, bs)
}

func (es *EtcdStore) ListMesosAgents() ([]*types.MesosAgent, error) {
	nodes, err := es.list(keyAgent)
	if err != nil {
		log.Errorln("etcd ListMesosAgents error:", err)
		return nil, err
	}

	agents := make([]*types.MesosAgent, 0)
	for node := range nodes {
		agent, err := es.GetMesosAgent(node)
		if err != nil {
			log.Errorf("get agent error: %v", err)
			continue
		}

		agents = append(agents, agent)
	}

	return agents, nil
}
