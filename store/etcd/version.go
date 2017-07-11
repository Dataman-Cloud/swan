package etcd

import (
	"path"

	log "github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *EtcdStore) CreateVersion(aid string, version *types.Version) error {
	bs, err := encode(version)
	if err != nil {
		return err
	}

	p := path.Join(keyApp, aid, keyVersions, version.ID)

	return s.create(p, bs)
}

func (s *EtcdStore) GetVersion(aid, vid string) (*types.Version, error) {
	p := path.Join(keyApp, aid, keyVersions, vid)

	data, err := s.get(p)
	if err != nil {
		log.Errorf("find app %s version %s got error: %v", aid, vid, err)
		return nil, err
	}

	var ver types.Version
	if err := decode(data, &ver); err != nil {
		return nil, err
	}

	return &ver, nil

}

func (s *EtcdStore) ListVersions(aid string) ([]*types.Version, error) {
	p := path.Join(keyApp, aid, keyVersions)

	children, err := s.list(p)
	if err != nil {
		log.Errorf("get app %s children(versions) error: %v", aid, err)
		return nil, err
	}

	versions := make([]*types.Version, 0)
	for child := range children {
		p := path.Join(keyApp, aid, keyVersions, child)
		data, err := s.get(p)
		if err != nil {
			log.Errorf("get %s got error: %v", p, err)
			return nil, err
		}

		var ver *types.Version
		if err := decode(data, &ver); err != nil {
			log.Errorf("decode task %s got error: %v", aid, err)
			return nil, err
		}

		versions = append(versions, ver)
	}

	return versions, nil
}
