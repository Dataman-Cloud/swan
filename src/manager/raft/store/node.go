package store

import (
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
)

func withCreateNodesBucketIfNotExists(tx *bolt.Tx, ID string, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyNodes, []byte(ID))
	if err != nil {
		return err
	}

	return fn(bkt)
}

func getNodesBucket(tx *bolt.Tx) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyNodes)
}

func getNodeBucket(tx *bolt.Tx, ID string) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyNodes, []byte(ID))
}

func withNodeBucket(tx *bolt.Tx, ID string, fn func(bkt *bolt.Bucket) error) error {
	bkt := getNodeBucket(tx, ID)
	if bkt == nil {
		return ErrNodeUnknown
	}

	return fn(bkt)
}

func createNode(tx *bolt.Tx, node *types.Node) error {
	return withCreateNodesBucketIfNotExists(tx, node.ID, func(bkt *bolt.Bucket) error {
		p, err := node.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func updateNode(tx *bolt.Tx, node *types.Node) error {
	return withNodeBucket(tx, node.ID, func(bkt *bolt.Bucket) error {
		p, err := node.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func removeNode(tx *bolt.Tx, ID string) error {
	nodesBkt := getNodesBucket(tx)
	if nodesBkt == nil {
		return nil
	}

	return nodesBkt.DeleteBucket([]byte(ID))
}
