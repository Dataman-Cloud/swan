package state

import (
	"errors"
	"sync"

	rafttypes "github.com/Dataman-Cloud/swan/src/manager/raft/types"
	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

var instance *OfferAllocator
var once sync.Once

type OfferAllocator struct {
	PendingOfferSlots []*Slot
	BySlotId          map[string]*mesos.OfferID // record allocated offers that map slot

	pendingOfferRWLock sync.RWMutex
	allocatedOfferLock sync.Mutex
}

func OfferAllocatorInstance() *OfferAllocator {
	once.Do(func() {
		instance = &OfferAllocator{
			PendingOfferSlots:  make([]*Slot, 0),
			BySlotId:           make(map[string]*mesos.OfferID),
			pendingOfferRWLock: sync.RWMutex{},
			allocatedOfferLock: sync.Mutex{},
		}
	})

	return instance
}

func (allocator *OfferAllocator) PopNextPendingOffer() *Slot {
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

func (allocator *OfferAllocator) RemoveSlotFromPendingOfferQueue(slot *Slot) {
	slotIndex := -1

	for index, pendingOfferSlot := range allocator.PendingOfferSlots {
		if pendingOfferSlot.ID == slot.ID {
			slotIndex = index
			break
		}
	}

	if slotIndex == -1 {
		return
	}
	allocator.pendingOfferRWLock.Lock()
	defer allocator.pendingOfferRWLock.Unlock()

	allocator.PendingOfferSlots = append(allocator.PendingOfferSlots[:slotIndex], allocator.PendingOfferSlots[slotIndex+1:]...)
}

func (allocator *OfferAllocator) SetOfferSlotMap(offerID *mesos.OfferID, slot *Slot) {
	allocator.allocatedOfferLock.Lock()
	allocator.BySlotId[slot.ID] = offerID
	allocator.create(slot.ID, *offerID.Value)
	allocator.allocatedOfferLock.Unlock()
}

func (allocator *OfferAllocator) RemoveOfferSlotMapBySlot(slot *Slot) {
	allocator.allocatedOfferLock.Lock()
	delete(allocator.BySlotId, slot.ID)
	allocator.allocatedOfferLock.Unlock()
	allocator.remove(slot.ID)
}

func (allocator *OfferAllocator) RemoveOfferSlotMapByOfferId(offerId *mesos.OfferID) {
	key := ""
	for k, v := range allocator.BySlotId {
		if v.Value == offerId.Value {
			key = k
		}
	}

	// shortcut execution if not found
	if len(key) == 0 {
		return
	}

	allocator.allocatedOfferLock.Lock()
	delete(allocator.BySlotId, key)
	allocator.allocatedOfferLock.Unlock()
}

func (allocator *OfferAllocator) RetriveSlotIdWithOfferId(offerId *mesos.OfferID) (string, error) {
	key := ""
	for k, v := range allocator.BySlotId {
		if v.Value == offerId.Value {
			key = k
		}
	}

	if len(key) == 0 {
		return "", errors.New("not found")
	}

	return key, nil
}

func (allocator *OfferAllocator) RemoveSlot(slot *Slot) {
	allocator.RemoveSlotFromPendingOfferQueue(slot)
	allocator.RemoveOfferSlotMapBySlot(slot)
}

func (allocator *OfferAllocator) create(slotID, offerID string) {
	logrus.Debugf("create offer allocator item %s => %s", slotID, offerID)
	persistentStore.CreateOfferAllocatorItem(context.TODO(), &rafttypes.OfferAllocatorItem{OfferID: offerID, SlotID: slotID}, nil)
}

func (allocator *OfferAllocator) remove(slotId string) {
	logrus.Debugf("remove offer allocator item  %s", slotId)
	persistentStore.DeleteOfferAllocatorItem(context.TODO(), slotId, nil)
}
