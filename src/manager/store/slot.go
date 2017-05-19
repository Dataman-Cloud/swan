package store

// TODO
// As Nested Field of AppHolder, CreateSlot Require Transaction Lock
func (zk *ZKStore) CreateSlot(slot *Slot) error {
	if zk.GetSlot(slot.AppID, slot.ID) != nil {
		return errSlotAlreadyExists
	}

	holder := zk.GetAppHolder(slot.AppID)
	if holder == nil {
		return errAppNotFound
	}

	holder.Slots[slot.ID] = slot

	bs, err := encode(holder)
	if err != nil {
		return err
	}

	path := keyApp + "/" + slot.AppID
	return zk.createAll(path, bs)
}

func (zk *ZKStore) GetSlot(aid, sid string) *Slot {
	holder := zk.GetAppHolder(aid)
	if holder == nil {
		return nil
	}

	return holder.Slots[sid]
}

func (zk *ZKStore) ListSlots(aid string) []*Slot {
	holder := zk.GetAppHolder(aid)
	if holder == nil {
		return nil
	}

	ret := make([]*Slot, 0, len(holder.Slots))
	for _, slot := range holder.Slots {
		ret = append(ret, slot)
	}
	return ret
}

// TODO
// As Nested Field of AppHolder, UpdateSlot Require Transaction Lock
func (zk *ZKStore) UpdateSlot(aid, sid string, slot *Slot) error {
	if zk.GetSlot(aid, sid) == nil {
		return errSlotNotFound
	}

	holder := zk.GetAppHolder(aid)
	if holder == nil {
		return errAppNotFound
	}

	holder.Slots[sid] = slot

	bs, err := encode(holder)
	if err != nil {
		return err
	}

	path := keyApp + "/" + aid
	return zk.createAll(path, bs)
}

// TODO
// As Nested Field of AppHolder, DeleteSlot Require Transaction Lock
func (zk *ZKStore) DeleteSlot(aid, sid string) error {
	if zk.GetSlot(aid, sid) == nil {
		return errSlotNotFound
	}

	holder := zk.GetAppHolder(aid)
	if holder == nil {
		return errAppNotFound
	}

	delete(holder.Slots, sid)

	bs, err := encode(holder)
	if err != nil {
		return err
	}

	path := keyApp + "/" + aid
	return zk.createAll(path, bs)
}
