package store

import (
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
)

func withCreateTaskBucketIfNotExists(tx *bolt.Tx, appId, slotId, taskId string, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(appId),
		bucketKeySlots, []byte(slotId), bucketKeyTasks, []byte(taskId))
	if err != nil {
		return err
	}

	return fn(bkt)
}

func WithTaskBucket(tx *bolt.Tx, appId, slotId, taskId string, fn func(bkt *bolt.Bucket) error) error {
	bkt := GetTaskBucket(tx, appId, taskId, slotId)
	if bkt == nil {
		return ErrTaskUnknown
	}

	return fn(bkt)
}

func GetTasksBucket(tx *bolt.Tx, appId, slotId string) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(appId),
		bucketKeySlots, []byte(slotId), bucketKeyTasks)
}

func GetTaskBucket(tx *bolt.Tx, appId, slotId, taskId string) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(appId),
		bucketKeySlots, []byte(slotId), bucketKeyTasks, []byte(taskId))
}

func createTask(tx *bolt.Tx, task *types.Task) error {
	return withCreateTaskBucketIfNotExists(tx, task.AppId, task.SlotId, task.Id, func(bkt *bolt.Bucket) error {
		p, err := task.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func updateTask(tx *bolt.Tx, task *types.Task) error {
	return WithTaskBucket(tx, task.AppId, task.SlotId, task.Id, func(bkt *bolt.Bucket) error {
		p, err := task.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func removeTask(tx *bolt.Tx, appId, slotId, taskId string) error {
	tasksBkt := GetTasksBucket(tx, appId, slotId)
	if tasksBkt == nil {
		return nil
	}

	return tasksBkt.DeleteBucket([]byte(taskId))
}
