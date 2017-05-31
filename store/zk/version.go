package zk

import (
	"path"

	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
)

// As Nested Field of AppHolder, CreateVersion Require Transaction Lock
func (zk *ZKStore) CreateVersion(aid string, version *types.Version) error {
	bs, err := encode(version)
	if err != nil {
		return err
	}

	p := path.Join(keyApp, aid, "versions", version.ID)

	return zk.createAll(p, bs)
}

func (zk *ZKStore) GetVersion(aid, vid string) (*types.Version, error) {
	p := path.Join(keyApp, aid, "versions", vid)

	data, _, err := zk.get(p)
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

func (zk *ZKStore) ListVersions(aid string) ([]*types.Version, error) {
	p := path.Join(keyApp, aid, "versions")

	children, err := zk.list(p)
	if err != nil {
		log.Errorf("get app %s children(versions) error: %v", aid, err)
		return nil, err
	}

	versions := make([]*types.Version, 0)
	for _, child := range children {
		p := path.Join(keyApp, aid, "versions", child)
		data, _, err := zk.get(p)
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
