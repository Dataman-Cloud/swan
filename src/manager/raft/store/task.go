package store

import (
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
)

func withCreateTaskBucketIfNotExists(tx *bolt.Tx, appID, slotID, taskID string, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(appID),
		bucketKeySlots, []byte(slotID), bucketKeyTasks, []byte(taskID))
	if err != nil {
		return err
	}

	return fn(bkt)
}

func WithTaskBucket(tx *bolt.Tx, appID, slotID, taskID string, fn func(bkt *bolt.Bucket) error) error {
	bkt := GetTaskBucket(tx, appID, taskID, slotID)
	if bkt == nil {
		return ErrTaskUnknown
	}

	return fn(bkt)
}

func GetTasksBucket(tx *bolt.Tx, appID, slotID string) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(appID),
		bucketKeySlots, []byte(slotID), bucketKeyTasks)
}

func GetTaskBucket(tx *bolt.Tx, appID, slotID, taskID string) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyApps, []byte(appID),
		bucketKeySlots, []byte(slotID), bucketKeyTasks, []byte(taskID))
}

func createTask(tx *bolt.Tx, task *types.Task) error {
	return withCreateTaskBucketIfNotExists(tx, task.AppID, task.SlotID, task.ID, func(bkt *bolt.Bucket) error {
		p, err := task.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func updateTask(tx *bolt.Tx, task *types.Task) error {
	return WithTaskBucket(tx, task.AppID, task.SlotID, task.ID, func(bkt *bolt.Bucket) error {
		p, err := task.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func removeTask(tx *bolt.Tx, appID, slotID, taskID string) error {
	tasksBkt := GetTasksBucket(tx, appID, slotID)
	if tasksBkt == nil {
		return nil
	}

	return tasksBkt.DeleteBucket([]byte(taskID))
}
