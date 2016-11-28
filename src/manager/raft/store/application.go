package store

import (
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
)

func withCreateAppBucketIfNotExists(tx *bolt.Tx, id string, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(id))
	if err != nil {
		return err
	}

	return fn(bkt)
}

func getAppsBucket(tx *bolt.Tx) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyApps)
}

func putApp(tx *bolt.Tx, app *types.Application) error {
	return withCreateAppBucketIfNotExists(tx, app.ID, func(bkt *bolt.Bucket) error {
		p, err := proto.Marshal(app)
		if err != nil {
			return nil
		}

		return bkt.Put(bucketKeyData, p)
	})
}

func removeApp(tx *bolt.Tx, appId string) error {
	appsBkt := getAppsBucket(tx)
	if appsBkt == nil {
		return nil
	}

	return appsBkt.DeleteBucket([]byte(appId))
}
