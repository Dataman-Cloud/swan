package zk

import (
	"fmt"
	"path"

	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
)

func (zk *ZKStore) CreateCompose(cps *types.Compose) error {
	bs, err := encode(cps)
	if err != nil {
		return err
	}

	path := path.Join(keyCompose, cps.ID)
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

	path := path.Join(keyCompose, cps.ID)
	return zk.set(path, bs)
}

func (zk *ZKStore) GetCompose(id string) (*types.Compose, error) {
	p := path.Join(keyCompose, id)

	bs, _, err := zk.get(p)
	if err != nil {
		log.Errorf("find compose %s got error: %v", id, err)
		return nil, fmt.Errorf("find compose %s got error: %v", id, err)
	}

	var cps = new(types.Compose)
	if err := decode(bs, &cps); err != nil {
		return nil, err
	}

	return cps, nil
}

func (zk *ZKStore) ListComposes() ([]*types.Compose, error) {
	ret := make([]*types.Compose, 0, 0)

	nodes, err := zk.list(keyCompose)
	if err != nil {
		log.Errorln("zk ListComposes error:", err)
		return ret, err
	}

	for _, node := range nodes {
		bs, _, err := zk.get(path.Join(keyCompose, node))
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

func (zk *ZKStore) DeleteCompose(id string) error {
	cps, err := zk.GetCompose(id)
	if err != nil {
		return err
	}

	return zk.del(path.Join(keyCompose, cps.ID))
}
