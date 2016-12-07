package store

import (
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/boltdb/bolt"
)

func withCreateManagerBucketIfNotExists(tx *bolt.Tx, id string, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, bucketKeyVersions, bucketKeyManager, []byte(id))
	if err != nil {
		return err
	}

	return fn(bkt)
}

func WithManagerBucket(tx *bolt.Tx, id string, fn func(bkt *bolt.Bucket) error) error {
	bkt := GetManagerBucket(tx, id)
	if bkt == nil {
		return nil
	}

	return fn(bkt)
}

func GetManagerBucket(tx *bolt.Tx, id string) *bolt.Bucket {
	return getBucket(tx, bucketKeyVersions, bucketKeyManager, []byte(id))
}

func putManager(tx *bolt.Tx, manager *types.Manager) error {
	return withCreateManagerBucketIfNotExists(tx, manager.ID, func(bkt *bolt.Bucket) error {
		p, err := manager.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}
