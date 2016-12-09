package store

import "github.com/Dataman-Cloud/swan/src/types"

func (store *ManagerStore) SaveVersion(version *types.Version) error {
	return nil
}

func (store *ManagerStore) ListVersionId(appId string) ([]string, error) {
	return nil, nil
}

func (store *ManagerStore) FetchVersion(appId, versionId string) (*types.Version, error) {
	return nil, nil
}

func (store *ManagerStore) DeleteVersion(appId, versionId string) error {
	return nil
}
