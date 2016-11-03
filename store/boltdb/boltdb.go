package boltdb

import (
	"errors"

	"github.com/boltdb/bolt"
)

type Boltdb struct {
	*bolt.DB
}

var (
	bucketKeyStorageVersion = []byte("v1")
	bucketKeyApps           = []byte("apps")
	bucketKeyData           = []byte("data")
	bucketKeyID             = []byte("ID")
	bucketKeyFramework      = []byte("framework")
)

var (
	errAppUnknown       = errors.New("boltdb: app unknown")
	errFrameworkUnknown = errors.New("boltdb: framework unknown")
)

func NewBoltdbStore(db *bolt.DB) *Boltdb {
	return &Boltdb{
		DB: db,
	}
}

func withCreateAppBucketIfNotExists(tx *bolt.Tx, id string, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(id))
	if err != nil {
		return err
	}

	return fn(bkt)
}

func createBucketIfNotExists(tx *bolt.Tx, keys ...[]byte) (*bolt.Bucket, error) {
	bkt, err := tx.CreateBucketIfNotExists(keys[0])
	if err != nil {
		return nil, err
	}

	for _, key := range keys[1:] {
		bkt, err = bkt.CreateBucketIfNotExists(key)
		if err != nil {
			return nil, err
		}
	}

	return bkt, nil
}

func withAppBucket(tx *bolt.Tx, id string, fn func(bkt *bolt.Bucket) error) error {
	bkt := getAppBucket(tx, id)
	if bkt == nil {
		return errAppUnknown
	}

	return fn(bkt)
}

func getAppBucket(tx *bolt.Tx, id string) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(id))
}

func getAppsBucket(tx *bolt.Tx) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyApps)
}

func getBucket(tx *bolt.Tx, keys ...[]byte) *bolt.Bucket {
	bkt := tx.Bucket(keys[0])

	for _, key := range keys[1:] {
		if bkt == nil {
			break
		}

		bkt = bkt.Bucket(key)
	}

	return bkt
}
