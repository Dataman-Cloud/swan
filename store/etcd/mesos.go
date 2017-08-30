package etcd

import (
	"fmt"
	"path"

	"github.com/Dataman-Cloud/swan/types"
)

func (zk *EtcdStore) CreateMesosAgent(agent *types.MesosAgent) error {
	p := path.Join(keyAgent, agent.ID)

	bs, err := encode(agent)
	if err != nil {
		return err
	}

	return zk.create(p, bs)
}

func (zk *EtcdStore) GetMesosAgent(agentId string) (*types.MesosAgent, error) {
	p := path.Join(keyAgent, agentId)

	data, err := zk.get(p)
	if err != nil {
		if zk.IsErrNotFound(err) {
			return nil, fmt.Errorf("agent %s not found", agentId)
		}

		return nil, err
	}

	var agent types.MesosAgent
	if err := decode(data, &agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

func (zk *EtcdStore) UpdateMesosAgent(agent *types.MesosAgent) error {
	bs, err := encode(agent)
	if err != nil {
		return err
	}

	p := path.Join(keyAgent, agent.ID, "value")

	return zk.update(p, bs)
}
