package store

// TODO
// As Nested Field of AppHolder, UpdateCurrentTask Require Transaction Lock
func (zk *ZKStore) UpdateCurrentTask(aid, sid string, task *Task) error {
	holder := zk.GetAppHolder(aid)
	if holder == nil {
		return errAppNotFound
	}

	slot := holder.Slots[sid]
	if slot == nil {
		return errSlotNotFound
	}

	slot.CurrentTask = task
	holder.Slots[sid] = slot

	bs, err := encode(holder)
	if err != nil {
		return err
	}

	path := keyApp + "/" + aid
	return zk.createAll(path, bs)
}

func (zk *ZKStore) ListTaskHistory(aid, sid string) []*Task {
	holder := zk.GetAppHolder(aid)
	if holder == nil {
		return nil
	}

	slot := holder.Slots[sid]
	if slot == nil {
		return nil
	}

	return slot.TaskHistory
}
