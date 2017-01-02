package store

import (
	raftstore "github.com/Dataman-Cloud/swan/src/manager/raft/store"
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
	"golang.org/x/net/context"
)

// as same as update version we need do this follow steps in one transaction
// 1. find this old slot info from data.
// 2. push this old slot's current task to task history.
// 3. set the new task as the slot's current task.
// 4. store the new slot info.
// 5. put all actions in one storeActions tp propose data.
func (s *FrameworkStore) UpdateTask(ctx context.Context, task *types.Task, cb func()) error {
	slot, err := s.GetSlot(task.AppID, task.SlotID)
	if err != nil {
		return err
	}

	if slot == nil {
		return ErrTaskNotFound
	}

	var storeActions []*types.StoreAction
	updateTaskAction := &types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Task{slot.CurrentTask},
	}
	storeActions = append(storeActions, updateTaskAction)

	slot.CurrentTask = task
	updateSlotAction := &types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Slot{slot},
	}
	storeActions = append(storeActions, updateSlotAction)

	return s.RaftNode.ProposeValue(ctx, storeActions, cb)
}

func (s *FrameworkStore) ListTasks(appId, slotId string) ([]*types.Task, error) {
	var tasks []*types.Task

	if err := s.BoltbDb.View(func(tx *bolt.Tx) error {
		bkt := raftstore.GetTasksBucket(tx, appId, slotId)
		if bkt == nil {
			tasks = []*types.Task{}
			return nil
		}

		return bkt.ForEach(func(k, v []byte) error {
			taskBkt := raftstore.GetTaskBucket(tx, appId, slotId, string(k))
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
