package boltdb

import (
	"github.com/Dataman-Cloud/swan/types"

	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
)

func (db *Boltdb) PutVersion(appId string, version *types.ApplicationVersion) error {
	return db.PutVersions(appId, version)
}

func (db *Boltdb) PutVersions(appId string, versions ...*types.ApplicationVersion) error {
	return db.Update(func(tx *bolt.Tx) error {
		for _, version := range versions {
			if err := withCreateVersionBucketIfNotExists(tx, appId, version.ID, func(bkt *bolt.Bucket) error {
				p, err := proto.Marshal(version)
				if err != nil {
					return err
				}

				return bkt.Put(bucketKeyData, p)
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func (db *Boltdb) GetVersion(appId, versionId string) (*types.ApplicationVersion, error) {
	versions, err := db.GetVersions(appId, versionId)
	if err != nil {
		return nil, err
	}

	if len(versions) < 1 || versions[0] == nil {
		return nil, errVersionUnknown
	}

	return versions[0], nil
}

func (db *Boltdb) GetVersions(appId string, versionIds ...string) ([]*types.ApplicationVersion, error) {
	if versionIds == nil {
		return db.getAllVersions(appId)
	}

	var versions []*types.ApplicationVersion

	if err := db.View(func(tx *bolt.Tx) error {
		for _, versionId := range versionIds {
			bkt := getVersionBucket(tx, appId, versionId)
			if bkt == nil {
				return nil
			}

			p := bkt.Get(bucketKeyData)

			var version types.ApplicationVersion
			if err := proto.Unmarshal(p, &version); err != nil {
				return err
			}

			versions = append(versions, &version)
		}

		return nil

	}); err != nil {
		return nil, err
	}

	return versions, nil
}

func (db *Boltdb) getAllVersions(appId string) ([]*types.ApplicationVersion, error) {
	var versions []*types.ApplicationVersion

	if err := db.View(func(tx *bolt.Tx) error {
		bkt := getVersionsBucket(tx, appId)
		if bkt == nil {
			versions = []*types.ApplicationVersion{}
			return nil
		}

		if err := bkt.ForEach(func(k, v []byte) error {
			versionBkt := bkt.Bucket(k)
			if versionBkt == nil {
				return nil
			}

			p := versionBkt.Get(bucketKeyData)

			var version types.ApplicationVersion
			if err := proto.Unmarshal(p, &version); err != nil {
				return err
			}

			versions = append(versions, &version)
			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return versions, nil
}

func (db *Boltdb) GetAndSortVersions(appId string, versionIds ...string) ([]*types.ApplicationVersion, error) {
	if versionIds == nil {
		return db.getAllAndSortVersions(appId)
	}

	var versions []*types.ApplicationVersion

	if err := db.View(func(tx *bolt.Tx) error {
		for _, versionId := range versionIds {
			bkt := getVersionBucket(tx, appId, versionId)
			if bkt == nil {
				return nil
			}

			p := bkt.Get(bucketKeyData)

			var version types.ApplicationVersion
			if err := proto.Unmarshal(p, &version); err != nil {
				return err
			}

			insertVersionById(versions, &version)
		}

		return nil

	}); err != nil {
		return nil, err
	}

	return versions, nil
}

func (db *Boltdb) getAllAndSortVersions(appId string) ([]*types.ApplicationVersion, error) {
	var versions []*types.ApplicationVersion

	if err := db.View(func(tx *bolt.Tx) error {
		bkt := getVersionsBucket(tx, appId)
		if bkt == nil {
			versions = []*types.ApplicationVersion{}
			return nil
		}

		if err := bkt.ForEach(func(k, v []byte) error {
			versionBkt := bkt.Bucket(k)
			if versionBkt == nil {
				return nil
			}

			p := versionBkt.Get(bucketKeyData)

			var version types.ApplicationVersion
			if err := proto.Unmarshal(p, &version); err != nil {
				return err
			}

			insertVersionById(versions, &version)
			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return versions, nil
}

func insertVersionById(versions []*types.ApplicationVersion, version *types.ApplicationVersion) []*types.ApplicationVersion {
	var sortedVersions []*types.ApplicationVersion
	if versions == nil || len(versions) == 0 {
		sortedVersions = append(versions, version)
		return sortedVersions
	}

	sortedVersions = append(versions, version)
	for i := len(sortedVersions) - 1; i > 0; i-- {
		if sortedVersions[i].ID > sortedVersions[i+1].ID {
			sortedVersions[i], sortedVersions[i+1] = sortedVersions[i+1], sortedVersions[i]
		} else {
			break
		}
	}

	return sortedVersions
}

func (db *Boltdb) DeleteVersion(appId, versionId string) error {
	return db.DeleteVersions(appId, versionId)
}

func (db *Boltdb) DeleteVersions(appId string, versionIds ...string) error {
	if versionIds == nil {
		return db.deleteAllVersions(appId)
	}

	return db.Update(func(tx *bolt.Tx) error {
		bkt := getVersionsBucket(tx, appId)
		if bkt == nil {
			return nil
		}

		for _, versionId := range versionIds {
			if err := bkt.DeleteBucket([]byte(versionId)); err != nil {
				return err
			}
		}

		return nil
	})
}

func (db *Boltdb) deleteAllVersions(appId string) error {
	return db.Update(func(tx *bolt.Tx) error {
		appBkt := getAppBucket(tx, appId)

		if appBkt == nil {
			return nil
		}

		return appBkt.DeleteBucket(bucketKeyVersions)
	})
}
