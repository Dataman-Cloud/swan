package store

import (
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
)

func withCreateQuotaBucketIfNotExists(tx *bolt.Tx, quotaGroup string, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyQuotas, []byte(quotaGroup))
	if err != nil {
		return err
	}

	return fn(bkt)
}

func WithQuotaBucket(tx *bolt.Tx, quotaGroup string, fn func(bkt *bolt.Bucket) error) error {
	bkt := GetQuotaBucket(tx, quotaGroup)
	if bkt == nil {
		return ErrQuotaUnknown
	}

	return fn(bkt)
}

func GetQuotaBucket(tx *bolt.Tx, quotaGroup string) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyQuotas, []byte(quotaGroup))
}

func GetQuotasBucket(tx *bolt.Tx) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyQuotas)
}

func createQuota(tx *bolt.Tx, quota *types.ResourceQuota) error {
	return withCreateQuotaBucketIfNotExists(tx, quota.QuotaGroup, func(bkt *bolt.Bucket) error {
		p, err := quota.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func updateQuota(tx *bolt.Tx, quota *types.ResourceQuota) error {
	return WithQuotaBucket(tx, quota.QuotaGroup, func(bkt *bolt.Bucket) error {
		p, err := quota.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func removeQuota(tx *bolt.Tx, quotaGroup string) error {
	quotasBkt := GetQuotasBucket(tx)
	if quotasBkt == nil {
		return nil
	}

	return quotasBkt.DeleteBucket([]byte(quotaGroup))
}
