package store

// As Nested Field of AppHolder, CreateVersion Require Transaction Lock
func (zk *ZKStore) CreateVersion(aid string, version *Version) error {
	zk.Lock()
	defer zk.Unlock()

	if zk.GetVersion(aid, version.ID) != nil {
		return errVersionAlreadyExists
	}

	holder := zk.GetAppHolder(aid)
	if holder == nil {
		return errAppNotFound
	}

	holder.Versions[version.ID] = version

	bs, err := encode(holder)
	if err != nil {
		return err
	}

	path := keyApp + "/" + aid
	return zk.createAll(path, bs)
}

func (zk *ZKStore) GetVersion(aid, vid string) *Version {
	holder := zk.GetAppHolder(aid)
	if holder == nil {
		return nil
	}

	return holder.Versions[vid]
}

func (zk *ZKStore) ListVersions(aid string) []*Version {
	holder := zk.GetAppHolder(aid)
	if holder == nil {
		return nil
	}

	ret := make([]*Version, 0, len(holder.Versions))
	for _, ver := range holder.Versions {
		ret = append(ret, ver)
	}
	return ret
}
