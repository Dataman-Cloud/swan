package store

func (zk *ZkStore) CreateSlot(slot *Slot) error {
	if zk.GetSlot(slot.AppID, slot.ID) != nil {
		return ErrSlotAlreadyExists
	}

	op := &AtomicOp{
		Op:      OP_ADD,
		Entity:  ENTITY_SLOT,
		Param1:  slot.AppID,
		Param2:  slot.ID,
		Payload: slot,
	}

	return zk.Apply(op, true)
}

func (zk *ZkStore) GetSlot(appId, slotId string) *Slot {
	appStore, found := zk.Storage.Apps[appId]
	if !found {
		return nil
	}

	slot, found := appStore.Slots[slotId]
	if !found {
		return nil
	}

	return slot.Slot
}

func (zk *ZkStore) ListSlots(appId string) []*Slot {
	appStore, found := zk.Storage.Apps[appId]
	if !found {
		return nil
	}

	slots := make([]*Slot, 0)
	for _, slotStore := range appStore.Slots {
		slots = append(slots, slotStore.Slot)
	}

	return slots
}

func (zk *ZkStore) UpdateSlot(appId, slotId string, slot *Slot) error {
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
		Entity:  ENTITY_SLOT,
		Param1:  slot.AppID,
		Param2:  slot.ID,
		Payload: slot,
	}

	return zk.Apply(op, true)
}

func (zk *ZkStore) DeleteSlot(appId, slotId string) error {
	appStore, found := zk.Storage.Apps[appId]
	if !found {
		return ErrAppNotFound
	}

	_, found = appStore.Slots[slotId]
	if !found {
		return ErrSlotNotFound
	}

	op := &AtomicOp{
		Op:     OP_REMOVE,
		Entity: ENTITY_SLOT,
		Param1: appId,
		Param2: slotId,
	}

	return zk.Apply(op, true)
}
