package store

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var (
	ZK_DEFAULT_ACL = zk.WorldACL(zk.PermAll)
)

const (
	OP_ADD = iota << 1
	OP_REMOVE
	OP_UPDATE
)

const (
	ENTITY_APP = 1
	ENTITY_SLOT
	ENTITY_VERSION
	ENTITY_CURRENT_TASK
	ENTITY_FRAMEWORKID
	ENTITY_OFFER_ALLOCATOR_ITEM
)

var (
	ErrAppNotFound  = errors.New("app not found")
	ErrSlotNotFound = errors.New("slot not found")
)

type StoreOp struct {
	Op     uint8
	Entity uint8
	Param1 string
	Param2 string
	Param3 string

	ZkPath  string
	Payload interface{}
}

type slotStorage struct {
	Slot        *Slot
	CurrentTask *Task
	TaskHistory []*Task
}

type appStorage struct {
	App      *Application
	Versions map[string]*Version
	Slots    map[string]slotStorage
}

type ZkStore struct {
	Apps           map[string]*appStorage
	OfferAllocator map[string]*OfferAllocatorItem
	FrameworkId    string

	mu   sync.RWMutex
	conn *zk.Conn
}

func NewZkStore() *ZkStore {
	conn, _, err := zk.Connect([]string{"114.55.130.152:2181"}, 5*time.Second)
	if err != nil {
		panic(err)
	}

	return &ZkStore{
		conn:           conn,
		Apps:           make(map[string]*appStorage),
		OfferAllocator: make(map[string]*OfferAllocatorItem),
	}
}

func (zkStore *ZkStore) Apply(op *StoreOp) error {
	//zk.Storage.Apply(op)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(op)
	if err != nil {
		return err
	}

	fmt.Println(op.Op, "  ", op.Entity)
	return nil
	//op.ZkPath, err = zkStore.conn.Create("/foobar-test/fsysy", buf.Bytes(), zk.FlagSequence, ZK_DEFAULT_ACL)
	//return err
}

func (zk *ZkStore) CreateVersion(appId string, version *Version) error {
	op := &StoreOp{
		Op:      OP_REMOVE,
		Entity:  ENTITY_APP,
		Param1:  appId,
		Param2:  version.ID,
		Payload: version,
	}

	return zk.Apply(op)
}

func (zk *ZkStore) GetVersion(appId, versionId string) *Version {
	appStore, found := zk.Apps[appId]
	if !found {
		return nil
	}

	version, found := appStore.Versions[versionId]
	if !found {
		return nil
	}

	return version
}

func (zk *ZkStore) ListVersions(appId string) []*Version {
	appStore, found := zk.Apps[appId]
	if !found {
		return nil
	}

	versions := make([]*Version, 0)
	for _, version := range appStore.Versions {
		versions = append(versions, version)
	}

	return versions
}

func (zk *ZkStore) CreateSlot(slot *Slot) error {
	op := &StoreOp{
		Op:      OP_ADD,
		Entity:  ENTITY_SLOT,
		Param1:  slot.AppID,
		Param2:  slot.ID,
		Payload: slot,
	}

	return zk.Apply(op)
}

func (zk *ZkStore) GetSlot(appId, slotId string) *Slot {
	appStore, found := zk.Apps[appId]
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
	appStore, found := zk.Apps[appId]
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
	appStore, found := zk.Apps[appId]
	if !found {
		return ErrSlotNotFound
	}

	_, found = appStore.Slots[slotId]
	if !found {
		return ErrSlotNotFound
	}

	op := &StoreOp{
		Op:      OP_UPDATE,
		Entity:  ENTITY_SLOT,
		Param1:  slot.AppID,
		Param2:  slot.ID,
		Payload: slot,
	}

	return zk.Apply(op)
}

func (zk *ZkStore) DeleteSlot(appId, slotId string) error {
	appStore, found := zk.Apps[appId]
	if !found {
		return ErrSlotNotFound
	}

	_, found = appStore.Slots[slotId]
	if !found {
		return ErrSlotNotFound
	}

	op := &StoreOp{
		Op:     OP_REMOVE,
		Entity: ENTITY_SLOT,
		Param1: appId,
		Param2: slotId,
	}

	return zk.Apply(op)
}

func (zk *ZkStore) UpdateCurrentTask(appId, slotId string, task *Task) error {
	appStore, found := zk.Apps[appId]
	if !found {
		return ErrSlotNotFound
	}

	_, found = appStore.Slots[slotId]
	if !found {
		return ErrSlotNotFound
	}

	op := &StoreOp{
		Op:      OP_UPDATE,
		Entity:  ENTITY_CURRENT_TASK,
		Param1:  appId,
		Param2:  slotId,
		Payload: task,
	}

	return zk.Apply(op)
}

func (zk *ZkStore) ListTaskHistory(appId, slotId string) []*Task {
	appStore, found := zk.Apps[appId]
	if !found {
		return nil
	}

	slotStore, found := appStore.Slots[slotId]
	if !found {
		return nil
	}

	return slotStore.TaskHistory
}

func (zk *ZkStore) UpdateFrameworkId(frameworkId string) error {
	op := &StoreOp{
		Op:      OP_UPDATE,
		Entity:  ENTITY_FRAMEWORKID,
		Payload: frameworkId,
	}

	return zk.Apply(op)
}

func (zk *ZkStore) GetFrameworkId() string {
	return zk.FrameworkId
}

func (zk *ZkStore) CreateOfferAllocatorItem(item *OfferAllocatorItem) error {
	op := &StoreOp{
		Op:      OP_ADD,
		Entity:  ENTITY_OFFER_ALLOCATOR_ITEM,
		Param1:  item.OfferID,
		Payload: item,
	}

	return zk.Apply(op)
}

func (zk *ZkStore) DeleteOfferAllocatorItem(offerId string) error {
	op := &StoreOp{
		Op:     OP_REMOVE,
		Entity: ENTITY_OFFER_ALLOCATOR_ITEM,
		Param1: offerId,
	}

	return zk.Apply(op)
}

func (zk *ZkStore) ListOfferallocatorItems() []*OfferAllocatorItem {
	items := make([]*OfferAllocatorItem, 0)
	for _, item := range zk.OfferAllocator {
		items = append(items, item)
	}

	return items
}
