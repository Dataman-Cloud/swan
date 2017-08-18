package etcd

import (
	log "github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *EtcdStore) CreateComposeNG(cmpApp *types.ComposeApp) error {
	bs, err := encode(cmpApp)
	if err != nil {
		return err
	}

	path := keyComposeNG + "/" + cmpApp.ID
	return s.create(path, bs)
}

func (s *EtcdStore) UpdateComposeNG(cmpApp *types.ComposeApp) error {
	if cmpApp, _ := s.GetComposeNG(cmpApp.ID); cmpApp == nil {
		return errComposeNotFound
	}

	bs, err := encode(cmpApp)
	if err != nil {
		return err
	}

	path := keyComposeNG + "/" + cmpApp.ID
	return s.update(path, bs)
}

func (s *EtcdStore) GetComposeNG(idOrName string) (*types.ComposeApp, error) {
	// by id
	bs, err := s.get(keyComposeNG + "/" + idOrName)
	if err == nil {
		cmpApp := new(types.ComposeApp)
		if err := decode(bs, &cmpApp); err != nil {
			log.Errorln("etcd GetComposeNG.decode error:", err)
			return nil, err
		}
		return cmpApp, nil
	}

	// by name
	cmpApps, err := s.ListComposesNG()
	if err != nil {
		return nil, err
	}
	for _, cmpApp := range cmpApps {
		if cmpApp.Name == idOrName {
			return cmpApp, nil
		}
	}

	return nil, errComposeNotFound
}

func (s *EtcdStore) ListComposesNG() ([]*types.ComposeApp, error) {
	ret := make([]*types.ComposeApp, 0, 0)

	nodes, err := s.list(keyComposeNG)
	if err != nil {
		log.Errorln("etcd ListComposesNG error:", err)
		return ret, err
	}

	for node := range nodes {
		bs, err := s.get(keyComposeNG + "/" + node)
		if err != nil {
			log.Errorln("etcd ListComposeNG.getnode error:", err)
			continue
		}

		cmpApp := new(types.ComposeApp)
		if err := decode(bs, &cmpApp); err != nil {
			log.Errorln("etcd ListComposeNG.decode error:", err)
			continue
		}

		ret = append(ret, cmpApp)
	}

	return ret, nil
}

func (s *EtcdStore) DeleteComposeNG(idOrName string) error {
	cmpApp, err := s.GetComposeNG(idOrName)
	if err != nil {
		return err
	}

	return s.delDir(keyComposeNG+"/"+cmpApp.ID, true)
}
