package state

import (
	"errors"
	"sync"

	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
)

type OfferAllocator struct {
	PendingOfferSlots []*Slot
	AllocatedOffer    map[string]*mesos.OfferID // record allocated offers that map slot

	pendingOfferWriteLock sync.Mutex
	allocatedOfferLock    sync.Mutex
}

func NewOfferAllocator() *OfferAllocator {
	allocator := &OfferAllocator{
		PendingOfferSlots:     make([]*Slot, 0),
		AllocatedOffer:        make(map[string]*mesos.OfferID),
		pendingOfferWriteLock: sync.Mutex{},
		allocatedOfferLock:    sync.Mutex{},
	}

	return allocator
}

func (allocator *OfferAllocator) NextPendingOffer() *Slot {
	if len(allocator.PendingOfferSlots) == 0 {
		return nil
	}

	slot := allocator.PendingOfferSlots[0]
	allocator.pendingOfferWriteLock.Lock()
	pendingOfferSize := len(allocator.PendingOfferSlots)
	allocator.PendingOfferSlots = allocator.PendingOfferSlots[1:pendingOfferSize]
	allocator.pendingOfferWriteLock.Unlock()

	return slot
}

func (allocator *OfferAllocator) PutSlotBackToPendingQueue(slot *Slot) {
	allocator.pendingOfferWriteLock.Lock()
	allocator.PendingOfferSlots = append(allocator.PendingOfferSlots, slot)
	allocator.pendingOfferWriteLock.Unlock()
}

func (allocator *OfferAllocator) SetOfferIdForSlotId(offerId *mesos.OfferID, slotName string) {
	allocator.allocatedOfferLock.Lock()
	allocator.AllocatedOffer[slotName] = offerId
	allocator.allocatedOfferLock.Unlock()
}

func (allocator *OfferAllocator) DeleteSlotIdOfferIdMap(offerId *mesos.OfferID) {
	key := ""
	for k, v := range allocator.AllocatedOffer {
		if v.Value == offerId.Value {
			key = k
		}
	}

	allocator.allocatedOfferLock.Lock()
	delete(allocator.AllocatedOffer, key)
	allocator.allocatedOfferLock.Unlock()
}

func (allocator *OfferAllocator) RetriveSlotIdOfferId(offerId *mesos.OfferID) (string, error) {
	key := ""
	for k, v := range allocator.AllocatedOffer {
		if v.Value == offerId.Value {
			key = k
		}
	}

	if len(key) == 0 {
		return "", errors.New("not found")
	}

	return key, nil
}
