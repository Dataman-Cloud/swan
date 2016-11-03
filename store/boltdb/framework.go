package boltdb

import "github.com/boltdb/bolt"

func (db *Boltdb) PutFrameworkID(id string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyFramework)
		if err != nil {
			return err
		}

		return bkt.Put(bucketKeyID, []byte(id))
	})
}

func (db *Boltdb) GetFrameworkID() (string, error) {
	var id string
	err := db.View(func(tx *bolt.Tx) error {
		bkt := getBucket(tx, bucketKeyStorageVersion, bucketKeyFramework)
		if bkt == nil {
			return errFrameworkUnknown
		}

		val := bkt.Get(bucketKeyID)
		id = string(val)
		return nil
	})

	return id, err
}
