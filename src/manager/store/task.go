package store

import (
	"errors"

	raftstore "github.com/Dataman-Cloud/swan/src/manager/raft/store"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/boltdb/bolt"
	"golang.org/x/net/context"
)

func (store *ManagerStore) SaveTask(task *types.Task) error {
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Task{task},
	}}

	return store.RaftNode.ProposeValue(context.TODO(), storeActions, nil)
}

func (store *ManagerStore) ListTasks(appId string) ([]*types.Task, error) {
	var tasks []*types.Task

	if err := store.BoltbDb.View(func(tx *bolt.Tx) error {
		bkt := raftstore.GetTasksBucket(tx, appId)
		if bkt == nil {
			tasks = []*types.Task{}
			return nil
		}

		return bkt.ForEach(func(k, v []byte) error {
			taskBkt := raftstore.GetTaskBucket(tx, appId, string(k))
			if taskBkt == nil {
				return nil
			}

			task := &types.Task{}
			p := taskBkt.Get(raftstore.BucketKeyData)
			if err := task.Unmarshal(p); err != nil {
				return err
			}

			tasks = append(tasks, task)
			return nil
		})

	}); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (store *ManagerStore) DeleteTask(appId, taskId string) error {
	task := &types.Task{Name: taskId, AppId: appId}
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_Task{task},
	}}

	return store.RaftNode.ProposeValue(context.TODO(), storeActions, nil)
}

func (store *ManagerStore) FetchTask(appId, taskId string) (*types.Task, error) {
	task := &types.Task{}

	if err := store.BoltbDb.View(func(tx *bolt.Tx) error {
		return raftstore.WithTaskBucket(tx, appId, taskId, func(bkt *bolt.Bucket) error {
			p := bkt.Get(raftstore.BucketKeyData)

			return task.Unmarshal(p)
		})
	}); err != nil {
		return nil, err
	}

	return task, nil
}

func (store *ManagerStore) UpdateTaskStatus(appId, taskId string, status string) error {
	task, err := store.FetchTask(appId, taskId)
	if err != nil {
		return err
	}

	if task == nil {
		return errors.New("task not found")
	}

	task.Status = status

	if err := store.SaveTask(task); err != nil {
		return err
	}

	return nil
}
