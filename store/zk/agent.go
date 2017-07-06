package zk

import (
	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
)

func (zk *ZKStore) CreateAgent(agent *types.Agent) error {
	bs, err := encode(agent)
	if err != nil {
		return err
	}

	path := keyAgent + "/" + agent.ID
	return zk.createAll(path, bs)
}

func (zk *ZKStore) UpdateAgent(agent *types.Agent) error {
	if a, _ := zk.GetAgent(agent.ID); a == nil {
		return errAgentNotFound
	}

	bs, err := encode(agent)
	if err != nil {
		return err
	}

	path := keyAgent + "/" + agent.ID
	return zk.create(path, bs)
}

func (zk *ZKStore) GetAgent(id string) (*types.Agent, error) {
	bs, _, err := zk.get(keyAgent + "/" + id)
	if err != nil {
		return nil, err
	}

	a := new(types.Agent)
	if err := decode(bs, &a); err != nil {
		log.Errorln("zk GetAgent.decode error:", err)
		return nil, err
	}

	return a, nil
}

func (zk *ZKStore) ListAgents() ([]*types.Agent, error) {
	ret := make([]*types.Agent, 0, 0)

	nodes, err := zk.list(keyAgent)
	if err != nil {
		log.Errorln("zk listAgents error:", err)
		return ret, err
	}

	for _, node := range nodes {
		bs, _, err := zk.get(keyAgent + "/" + node)
		if err != nil {
			log.Errorln("zk ListAgents.getnode error:", err)
			continue
		}

		a := new(types.Agent)
		if err := decode(bs, &a); err != nil {
			log.Errorln("zk LisetAgents.decode error:", err)
			continue
		}

		ret = append(ret, a)
	}

	return ret, nil
}
