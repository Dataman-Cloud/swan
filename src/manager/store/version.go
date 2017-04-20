package store

func (zk *ZkStore) CreateVersion(appId string, version *Version) error {
	if zk.GetVersion(appId, version.ID) != nil {
		return ErrVersionAlreadyExists
	}

	op := &AtomicOp{
		Op:      OP_ADD,
		Entity:  ENTITY_VERSION,
		Param1:  appId,
		Param2:  version.ID,
		Payload: version,
	}

	return zk.Apply(op, true)
}

func (zk *ZkStore) GetVersion(appId, versionId string) *Version {
	zk.mu.RLock()
	defer zk.mu.RUnlock()

	appStore, found := zk.Storage.Apps[appId]
	if !found {
		return nil
	}

	version, found := appStore.Versions[versionId]
	if !found {
		return nil
	}

	return version
}

func (zk *ZkStore) ListVersions(appId string) []*Version {
	zk.mu.RLock()
	defer zk.mu.RUnlock()

	appStore, found := zk.Storage.Apps[appId]
	if !found {
		return nil
	}

	versions := make([]*Version, 0)
	for _, version := range appStore.Versions {
		versions = append(versions, version)
	}

	return versions
}
