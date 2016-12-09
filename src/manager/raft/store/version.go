package store

//
//import (
//	"github.com/Dataman-Cloud/swan/src/manager/raft/types"
//
//	"github.com/boltdb/bolt"
//)
//
//func withCreateVersionBucketIfNotExists(tx *bolt.Tx, appId, versionId string, fn func(bkt *bolt.Bucket) error) error {
//	bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(appId), bucketKeyVersions, []byte(versionId))
//	if err != nil {
//		return err
//	}
//
//	return fn(bkt)
//}
//
//func WithVersionBucket(tx *bolt.Tx, appId, versionId string, fn func(bkt *bolt.Bucket) error) error {
//	bkt := GetVersionBucket(tx, appId, versionId)
//	if bkt == nil {
//		return ErrVersionUnknown
//	}
//
//	return fn(bkt)
//}
//
//func GetVersionsBucket(tx *bolt.Tx, appId string) *bolt.Bucket {
//	return getBucket(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(appId), bucketKeyVersions)
//}
//
//func GetVersionBucket(tx *bolt.Tx, appId, versionId string) *bolt.Bucket {
//	return getBucket(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(appId), bucketKeyVersions, []byte(versionId))
//}
//
//func putVersion(tx *bolt.Tx, version *types.Version) error {
//	return withCreateVersionBucketIfNotExists(tx, version.AppId, version.ID, func(bkt *bolt.Bucket) error {
//		p, err := version.Marshal()
//		if err != nil {
//			return err
//		}
//
//		return bkt.Put(BucketKeyData, p)
//	})
//}
//
//func removeVersion(tx *bolt.Tx, appId, versionId string) error {
//	versionsBkt := GetVersionsBucket(tx, appId)
//	if versionsBkt == nil {
//		return nil
//	}
//
//	return versionsBkt.DeleteBucket([]byte(versionId))
//}
