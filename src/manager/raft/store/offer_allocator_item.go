package store

import (
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
)

func withCreateOfferAllocatorItemBucketIfNotExists(tx *bolt.Tx, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyOfferAllocator)
	if err != nil {
		return err
	}

	return fn(bkt)
}

func WithOfferAllocatorItemBucket(tx *bolt.Tx, fn func(bkt *bolt.Bucket) error) error {
	bkt := GetOfferAllocatorItemBucket(tx)
	if bkt == nil {
		return ErrSlotUnknown
	}

	return fn(bkt)
}

func GetOfferAllocatorItemBucket(tx *bolt.Tx) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyOfferAllocator)
}

func createOfferAllocatorItem(tx *bolt.Tx, item *types.OfferAllocatorItem) error {
	return withCreateOfferAllocatorItemBucketIfNotExists(tx, func(bkt *bolt.Bucket) error {
		p, err := item.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put([]byte(item.SlotID), p)
	})
}

func removeOfferAllocatorItem(tx *bolt.Tx, item *types.OfferAllocatorItem) error {
	bkt := GetOfferAllocatorItemBucket(tx)

	if bkt == nil {
		return nil
	}

	return bkt.Delete([]byte(item.SlotID))
}
