package store

import (
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
)

func withCreateVersionBucketIfNotExists(tx *bolt.Tx, appID, versionID string, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(appID), bucketKeyVersions, []byte(versionID))
	if err != nil {
		return err
	}

	return fn(bkt)
}

func WithVersionBucket(tx *bolt.Tx, appID, versionID string, fn func(bkt *bolt.Bucket) error) error {
	bkt := GetVersionBucket(tx, appID, versionID)
	if bkt == nil {
		return ErrVersionUnknown
	}

	return fn(bkt)
}

func GetVersionsBucket(tx *bolt.Tx, appID string) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(appID), bucketKeyVersions)
}

func GetVersionBucket(tx *bolt.Tx, appID, versionID string) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(appID), bucketKeyVersions, []byte(versionID))
}

func createVersion(tx *bolt.Tx, version *types.Version) error {
	return withCreateVersionBucketIfNotExists(tx, version.AppID, version.ID, func(bkt *bolt.Bucket) error {
		p, err := version.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func updateVersion(tx *bolt.Tx, version *types.Version) error {
	return WithVersionBucket(tx, version.AppID, version.ID, func(bkt *bolt.Bucket) error {
		p, err := version.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func removeVersion(tx *bolt.Tx, appID, versionID string) error {
	versionsBkt := GetVersionsBucket(tx, appID)
	if versionsBkt == nil {
		return nil
	}

	return versionsBkt.DeleteBucket([]byte(versionID))
}
