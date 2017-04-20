package store

func (zk *ZkStore) UpdateCurrentTask(appId, slotId string, task *Task) error {
	appStore, found := zk.Storage.Apps[appId]
	if !found {
		return ErrAppNotFound
	}

	_, found = appStore.Slots[slotId]
	if !found {
		return ErrSlotNotFound
	}

	op := &AtomicOp{
		Op:      OP_UPDATE,
		Entity:  ENTITY_CURRENT_TASK,
		Param1:  appId,
		Param2:  slotId,
		Payload: task,
	}

	return zk.Apply(op, true)
}

func (zk *ZkStore) ListTaskHistory(appId, slotId string) []*Task {
	zk.mu.RLock()
	defer zk.mu.RUnlock()

	appStore, found := zk.Storage.Apps[appId]
	if !found {
		return nil
	}

	slotStore, found := appStore.Slots[slotId]
	if !found {
		return nil
	}

	return slotStore.TaskHistory
}
