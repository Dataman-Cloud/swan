package state

import (
	"errors"
	"sync"

	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
)

type OfferAllocator struct {
	PendingOfferSlots     []*Slot
	AllocatedOffer        map[string]*mesos.OfferID // record allocated offers that map slot
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

func (allocator *OfferAllocator) SetOfferIdForSlotName(offerId *mesos.OfferID, slotName string) {
	allocator.allocatedOfferLock.Lock()
	allocator.AllocatedOffer[slotName] = offerId
	allocator.allocatedOfferLock.Unlock()
}

func (allocator *OfferAllocator) DeleteSlotNameOfferIdMap(offerId *mesos.OfferID) {
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

func (allocator *OfferAllocator) RetriveSlotNameOfferId(offerId *mesos.OfferID) (string, error) {
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

// wrapper offer to record offer reserve history
type OfferWrapper struct {
	Offer    *mesos.Offer
	CpusUsed float64
	MemUsed  float64
	DiskUsed float64
	PortUsed []int
}

func NewOfferWrapper(offer *mesos.Offer) *OfferWrapper {
	o := &OfferWrapper{
		Offer:    offer,
		CpusUsed: 0,
		MemUsed:  0,
		DiskUsed: 0,
		PortUsed: make([]int, 0),
	}
	return o
}

func (ow *OfferWrapper) CpuRemain() float64 {
	var cpus float64
	for _, res := range ow.Offer.GetResources() {
		if res.GetName() == "cpus" {
			cpus += *res.GetScalar().Value
		}
	}

	return cpus - ow.CpusUsed
}

func (ow *OfferWrapper) MemRemain() float64 {
	var mem float64
	for _, res := range ow.Offer.GetResources() {
		if res.GetName() == "mem" {
			mem += *res.GetScalar().Value
		}
	}

	return mem - ow.MemUsed
}

func (ow *OfferWrapper) DiskRemain() float64 {
	var disk float64
	for _, res := range ow.Offer.GetResources() {
		if res.GetName() == "disk" {
			disk += *res.GetScalar().Value
		}
	}

	return disk - ow.DiskUsed
}
