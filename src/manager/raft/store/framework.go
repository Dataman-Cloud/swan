package store

import (
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
)

func withCreateFrameworkBucketIfNotExists(tx *bolt.Tx, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyFramework)
	if err != nil {
		return err
	}

	return fn(bkt)
}

func WithFrameworkBucket(tx *bolt.Tx, fn func(bkt *bolt.Bucket) error) error {
	bkt := GetFrameworkBucket(tx)
	if bkt == nil {
		return nil
	}

	return fn(bkt)
}

func GetFrameworkBucket(tx *bolt.Tx) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyFramework)
}

func putFramework(tx *bolt.Tx, framework *types.Framework) error {
	return withCreateFrameworkBucketIfNotExists(tx, func(bkt *bolt.Bucket) error {
		p, err := framework.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func removeFramework(tx *bolt.Tx) error {
	frameworkBkt := GetFrameworkBucket(tx)
	if frameworkBkt == nil {
		return nil
	}

	return tx.DeleteBucket(bucketKeyFramework)
}
