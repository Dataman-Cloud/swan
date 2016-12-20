package state

import (
	"errors"
	"sync"

	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
)

type OfferAllocator struct {
	PendingOfferSlots []*Slot
	BySlotName        map[string]*mesos.OfferID // record allocated offers that map slot

	pendingOfferRWLock sync.RWMutex
	allocatedOfferLock sync.Mutex
}

func NewOfferAllocator() *OfferAllocator {
	allocator := &OfferAllocator{
		PendingOfferSlots:  make([]*Slot, 0),
		BySlotName:         make(map[string]*mesos.OfferID),
		pendingOfferRWLock: sync.RWMutex{},
		allocatedOfferLock: sync.Mutex{},
	}

	return allocator
}

func (allocator *OfferAllocator) NextPendingOffer() *Slot {
	if len(allocator.PendingOfferSlots) == 0 {
		return nil
	}

	slot := allocator.PendingOfferSlots[0]
	allocator.pendingOfferRWLock.Lock()
	pendingOfferSize := len(allocator.PendingOfferSlots)
	allocator.PendingOfferSlots = allocator.PendingOfferSlots[1:pendingOfferSize]
	allocator.pendingOfferRWLock.Unlock()

	return slot
}

func (allocator *OfferAllocator) PutSlotBackToPendingQueue(slot *Slot) {
	allocator.pendingOfferRWLock.Lock()
	allocator.PendingOfferSlots = append(allocator.PendingOfferSlots, slot)
	allocator.pendingOfferRWLock.Unlock()
}

func (allocator *OfferAllocator) SetOfferIdForSlotId(offerId *mesos.OfferID, slotName string) {
	allocator.allocatedOfferLock.Lock()
	allocator.BySlotName[slotName] = offerId
	allocator.allocatedOfferLock.Unlock()
}

func (allocator *OfferAllocator) DeleteSlotIdOfferIdMap(offerId *mesos.OfferID) {
	key := ""
	for k, v := range allocator.BySlotName {
		if v.Value == offerId.Value {
			key = k
		}
	}

	allocator.allocatedOfferLock.Lock()
	delete(allocator.BySlotName, key)
	allocator.allocatedOfferLock.Unlock()
}

func (allocator *OfferAllocator) RetriveSlotIdOfferId(offerId *mesos.OfferID) (string, error) {
	key := ""
	for k, v := range allocator.BySlotName {
		if v.Value == offerId.Value {
			key = k
		}
	}

	if len(key) == 0 {
		return "", errors.New("not found")
	}

	return key, nil
}
