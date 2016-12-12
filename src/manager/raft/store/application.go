package store

import (
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
)

func withCreateAppBucketIfNotExists(tx *bolt.Tx, id string, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(id))
	if err != nil {
		return err
	}

	return fn(bkt)
}

func WithAppBucket(tx *bolt.Tx, id string, fn func(bkt *bolt.Bucket) error) error {
	bkt := GetAppBucket(tx, id)
	if bkt == nil {
		return ErrAppUnknown
	}

	return fn(bkt)
}

func GetAppBucket(tx *bolt.Tx, id string) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(id))
}

func GetAppsBucket(tx *bolt.Tx) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyApps)
}

func putApp(tx *bolt.Tx, app *types.Application) error {
	return withCreateAppBucketIfNotExists(tx, app.ID, func(bkt *bolt.Bucket) error {
		p, err := app.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func removeApp(tx *bolt.Tx, appId string) error {
	appsBkt := GetAppsBucket(tx)
	if appsBkt == nil {
		return nil
	}

	return appsBkt.DeleteBucket([]byte(appId))
}
