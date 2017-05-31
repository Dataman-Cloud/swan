package zk

import (
	"errors"

	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
)

func (zk *ZKStore) CreateCompose(cps *types.Compose) error {
	bs, err := encode(cps)
	if err != nil {
		return err
	}

	path := keyCompose + "/" + cps.ID
	return zk.createAll(path, bs)
}

func (zk *ZKStore) UpdateCompose(cps *types.Compose) error {
	if i, _ := zk.GetCompose(cps.ID); i == nil {
		return errInstanceNotFound
	}

	bs, err := encode(cps)
	if err != nil {
		return err
	}

	path := keyCompose + "/" + cps.ID
	return zk.create(path, bs)
}

func (zk *ZKStore) GetCompose(idOrName string) (*types.Compose, error) {
	// by id
	bs, _, err := zk.get(keyCompose + "/" + idOrName)
	if err == nil {
		cps := new(types.Compose)
		if err := decode(bs, &cps); err != nil {
			log.Errorln("zk GetCompose.decode error:", err)
			return nil, err
		}
		return cps, nil
	}

	// by name
	cpss, err := zk.ListComposes()
	if err != nil {
		return nil, err
	}
	for _, cps := range cpss {
		if cps.Name == idOrName {
			return cps, nil
		}
	}

	return nil, errors.New("no such compose")
}

func (zk *ZKStore) ListComposes() ([]*types.Compose, error) {
	ret := make([]*types.Compose, 0, 0)

	nodes, err := zk.list(keyCompose)
	if err != nil {
		log.Errorln("zk ListComposes error:", err)
		return ret, err
	}

	for _, node := range nodes {
		bs, _, err := zk.get(keyCompose + "/" + node)
		if err != nil {
			log.Errorln("zk ListCompose.getnode error:", err)
			continue
		}

		cps := new(types.Compose)
		if err := decode(bs, &cps); err != nil {
			log.Errorln("zk ListCompose.decode error:", err)
			continue
		}

		ret = append(ret, cps)
	}

	return ret, nil
}

func (zk *ZKStore) DeleteCompose(idOrName string) error {
	cps, err := zk.GetCompose(idOrName)
	if err != nil {
		return err
	}

	return zk.del(keyCompose + "/" + cps.ID)
}
