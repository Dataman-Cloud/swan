package etcd

import (
	log "github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *EtcdStore) CreateAgent(agent *types.Agent) error {
	bs, err := encode(agent)
	if err != nil {
		return err
	}

	path := keyAgent + "/" + agent.ID
	return s.create(path, bs)
}

func (s *EtcdStore) UpdateAgent(agent *types.Agent) error {
	if a, _ := s.GetAgent(agent.ID); a == nil {
		return errAgentNotFound
	}

	bs, err := encode(agent)
	if err != nil {
		return err
	}

	path := keyAgent + "/" + agent.ID
	return s.update(path, bs)
}

func (s *EtcdStore) GetAgent(id string) (*types.Agent, error) {
	bs, err := s.get(keyAgent + "/" + id)
	if err != nil {
		return nil, err
	}

	a := new(types.Agent)
	if err := decode(bs, &a); err != nil {
		log.Errorln("etcd GetAgent.decode error:", err)
		return nil, err
	}

	return a, nil
}

func (s *EtcdStore) ListAgents() ([]*types.Agent, error) {
	ret := make([]*types.Agent, 0, 0)

	nodes, err := s.list(keyAgent)
	if err != nil {
		log.Errorln("etcd listAgents error:", err)
		return ret, err
	}

	for id, node := range nodes {
		a := new(types.Agent)
		if err := decode(node, &a); err != nil {
			log.Errorln("etcd LisetAgents.decode error:", id, err)
			continue
		}

		ret = append(ret, a)
	}

	return ret, nil
}
