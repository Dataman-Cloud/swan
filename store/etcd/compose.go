package etcd

import (
	log "github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *EtcdStore) CreateCompose(cps *types.Compose) error {
	bs, err := encode(cps)
	if err != nil {
		return err
	}

	path := keyCompose + "/" + cps.ID
	return s.create(path, bs)
}

func (s *EtcdStore) UpdateCompose(cps *types.Compose) error {
	if i, _ := s.GetCompose(cps.ID); i == nil {
		return errComposeNotFound
	}

	bs, err := encode(cps)
	if err != nil {
		return err
	}

	path := keyCompose + "/" + cps.ID
	return s.update(path, bs)
}

func (s *EtcdStore) GetCompose(idOrName string) (*types.Compose, error) {
	// by id
	bs, err := s.get(keyCompose + "/" + idOrName)
	if err == nil {
		cps := new(types.Compose)
		if err := decode(bs, &cps); err != nil {
			log.Errorln("etcd GetCompose.decode error:", err)
			return nil, err
		}
		return cps, nil
	}

	// by name
	cpss, err := s.ListComposes()
	if err != nil {
		return nil, err
	}
	for _, cps := range cpss {
		if cps.Name == idOrName {
			return cps, nil
		}
	}

	return nil, errComposeNotFound
}

func (s *EtcdStore) ListComposes() ([]*types.Compose, error) {
	ret := make([]*types.Compose, 0, 0)

	nodes, err := s.list(keyCompose)
	if err != nil {
		log.Errorln("etcd ListComposes error:", err)
		return ret, err
	}

	for node := range nodes {
		bs, err := s.get(keyCompose + "/" + node)
		if err != nil {
			log.Errorln("etcd ListCompose.getnode error:", err)
			continue
		}

		cps := new(types.Compose)
		if err := decode(bs, &cps); err != nil {
			log.Errorln("etcd ListCompose.decode error:", err)
			continue
		}

		ret = append(ret, cps)
	}

	return ret, nil
}

func (s *EtcdStore) DeleteCompose(idOrName string) error {
	cps, err := s.GetCompose(idOrName)
	if err != nil {
		return err
	}

	return s.delDir(keyCompose+"/"+cps.ID, true)
}
