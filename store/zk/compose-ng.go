package zk

import (
	"path"

	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
)

func (zk *ZKStore) CreateComposeNG(cmpApp *types.ComposeApp) error {
	bs, err := encode(cmpApp)
	if err != nil {
		return err
	}

	path := path.Join(keyComposeNG, cmpApp.ID)
	return zk.createAll(path, bs)
}

func (zk *ZKStore) UpdateComposeNG(cmpApp *types.ComposeApp) error {
	if cmpApp, _ := zk.GetComposeNG(cmpApp.ID); cmpApp == nil {
		return errComposeNotFound
	}

	bs, err := encode(cmpApp)
	if err != nil {
		return err
	}

	path := path.Join(keyComposeNG, cmpApp.ID)
	return zk.set(path, bs)
}

func (zk *ZKStore) GetComposeNG(id string) (*types.ComposeApp, error) {
	// try id
	p := path.Join(keyComposeNG, id)

	bs, _, err := zk.get(p)
	if err == nil {
		var cps = new(types.ComposeApp)
		if err := decode(bs, &cps); err != nil {
			log.Errorf("zk GetComposeNG.decode() got error: %v", id, err)
			return nil, err
		}
		return cps, nil
	}

	// try name
	cmpApps, err := zk.ListComposesNG()
	if err != nil {
		log.Errorf("zk GetComposeNG.list() got error: %v", id, err)
		return nil, err
	}
	for _, cmpApp := range cmpApps {
		if cmpApp.Name == id {
			return cmpApp, nil
		}
	}

	return nil, errComposeNotFound
}

func (zk *ZKStore) ListComposesNG() ([]*types.ComposeApp, error) {
	ret := make([]*types.ComposeApp, 0, 0)

	nodes, err := zk.list(keyComposeNG)
	if err != nil {
		log.Errorln("zk ListComposesNG error:", err)
		return ret, err
	}

	for _, node := range nodes {
		bs, _, err := zk.get(path.Join(keyComposeNG, node))
		if err != nil {
			log.Errorln("zk ListComposeNG.getnode error:", err)
			continue
		}

		cps := new(types.ComposeApp)
		if err := decode(bs, &cps); err != nil {
			log.Errorln("zk ListComposeNG.decode error:", err)
			continue
		}

		ret = append(ret, cps)
	}

	return ret, nil
}

func (zk *ZKStore) DeleteComposeNG(id string) error {
	cmpApp, err := zk.GetComposeNG(id)
	if err != nil {
		return err
	}

	return zk.del(path.Join(keyComposeNG, cmpApp.ID))
}
