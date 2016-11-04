package boltdb

import (
	"github.com/Dataman-Cloud/swan/types"

	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
)

func (db *Boltdb) PutTasks(tasks ...*types.Task) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, task := range tasks {
		if err := withCreateTaskBucketIfNotExists(tx, task.AppId, task.ID, func(bkt *bolt.Bucket) error {
			p, err := proto.Marshal(task)
			if err != nil {
				return err
			}

			return bkt.Put(bucketKeyData, p)
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *Boltdb) PutTask(task *types.Task) error {
	return db.PutTasks(task)
}

func (db *Boltdb) UpdateTaskStatus(appId, taskId, status string) error {
	task, err := db.GetTask(appId, taskId)
	if err != nil {
		return err
	}

	task.Status = status
	return db.PutTask(task)
}

func (db *Boltdb) GetTask(appId, taskId string) (*types.Task, error) {
	tasks, err := db.GetTasks(appId, taskId)
	if err != nil {
		return nil, err
	}

	if len(tasks) < 1 {
		return nil, errTaskUnknown
	}

	return tasks[0], nil
}

func (db *Boltdb) GetTasks(appId string, taskIds ...string) ([]*types.Task, error) {
	if taskIds == nil {
		return db.getAllTasks(appId)
	}

	var tasks []*types.Task

	if err := db.View(func(tx *bolt.Tx) error {
		for _, taskId := range taskIds {
			bkt := getTaskBucket(tx, appId, taskId)
			if bkt == nil {
				return errTaskUnknown
			}

			p := bkt.Get(bucketKeyData)
			var task types.Task
			if err := proto.Unmarshal(p, &task); err != nil {
				return err
			}

			tasks = append(tasks, &task)
			return nil
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (db *Boltdb) getAllTasks(appId string) ([]*types.Task, error) {
	var tasks []*types.Task

	if err := db.View(func(tx *bolt.Tx) error {
		bkt := getTasksBucket(tx, appId)
		if bkt == nil {
			tasks = []*types.Task{}
			return nil
		}

		if err := bkt.ForEach(func(k, v []byte) error {
			taskBkt := bkt.Bucket(k)
			if taskBkt == nil {
				return nil
			}

			var task types.Task
			p := taskBkt.Get(bucketKeyData)
			if err := proto.Unmarshal(p, &task); err != nil {
				return err
			}

			tasks = append(tasks, &task)
			return nil
		}); err != nil {
			return err
		}
		return nil

	}); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (db *Boltdb) DeleteTask(appId, taskId string) error {
	return db.DeleteTasks(appId, taskId)
}

func (db *Boltdb) DeleteTasks(appId string, taskIds ...string) error {
	if taskIds == nil {
		return db.deleteAllTasks(appId)
	}

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bkt := getTasksBucket(tx, appId)
	if bkt == nil {
		return nil
	}

	for _, taskId := range taskIds {
		if err := bkt.DeleteBucket([]byte(taskId)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *Boltdb) deleteAllTasks(appId string) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bkt := getAppBucket(tx, appId)
	if bkt == nil {
		return nil
	}

	if err := bkt.DeleteBucket(bucketKeyTasks); err != nil {
		return err
	}
	return tx.Commit()
}
