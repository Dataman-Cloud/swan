package store

func (zk *ZkStore) CreateOfferAllocatorItem(item *OfferAllocatorItem) error {
	op := &AtomicOp{
		Op:      OP_ADD,
		Entity:  ENTITY_OFFER_ALLOCATOR_ITEM,
		Param1:  item.SlotID,
		Payload: item,
	}

	return zk.Apply(op, true)
}

func (zk *ZkStore) DeleteOfferAllocatorItem(offerId string) error {
	op := &AtomicOp{
		Op:     OP_REMOVE,
		Entity: ENTITY_OFFER_ALLOCATOR_ITEM,
		Param1: offerId,
	}

	return zk.Apply(op, true)
}

func (zk *ZkStore) ListOfferallocatorItems() []*OfferAllocatorItem {
	zk.mu.RLock()
	defer zk.mu.RUnlock()

	items := make([]*OfferAllocatorItem, 0)
	for _, item := range zk.Storage.OfferAllocator {
		items = append(items, item)
	}

	return items
}
