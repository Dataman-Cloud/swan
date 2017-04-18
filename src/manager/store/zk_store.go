package store

import (
	"bytes"
	"encoding/gob"
	"errors"
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
	ENTITY_TASK
	ENTITY_FRAMEWORK
	ENTITY_OFFER_ALLOCATOR_ITEM
)

var (
	ErrAppNotFound = errors.New("app not found")
)

type StoreOp struct {
	Op     uint8
	Entity uint8
	Param1 string
	Param2 string
	Param3 string

	ZkPath  string
	Payload []byte
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

	mu   sync.RWMutex
	conn *zk.Conn
}

func NewZkStore() *ZkStore {
	conn, _, err := zk.Connect([]string{"zk://114.55.130.152:2181/foobar-test"}, 5*time.Second)
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

	op.ZkPath, err = zkStore.conn.Create("/foobar-test/fsysy", buf.Bytes(), zk.FlagSequence, ZK_DEFAULT_ACL)
	return err
}

func (zk *ZkStore) CreateVersion(appId string, version *Version) error {
	op := &StoreOp{
		Op:     OP_REMOVE,
		Entity: ENTITY_APP,
		Param1: appId,
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
	return nil
}

func (zk *ZkStore) CreateSlot(slot *Slot) error {
	return nil
}

func (zk *ZkStore) GetSlot(appId, slotId string) *Slot {
	return nil
}

func (zk *ZkStore) ListSlots(appId string) []*Slot {
	return nil
}

func (zk *ZkStore) UpdateSlot(slot *Slot) error {
	return nil
}

func (zk *ZkStore) DeleteSlot(appId, slotId string) error {
	return nil
}

func (zk *ZkStore) UpdateTask(task *Task) error {
	return nil
}

func (zk *ZkStore) ListTasks(appId, slotId string) []*Task {
	return nil
}

func (zk *ZkStore) UpdateFrameworkId(frameworkId string) error {
	return nil
}

func (zk *ZkStore) GetFrameworkId() string {
	return ""
}

func (zk *ZkStore) CreateOfferAllocatorItem(item *OfferAllocatorItem) error {
	return nil
}

func (zk *ZkStore) DeleteOfferAllocatorItem(offerId string) error {
	return nil
}

func (zk *ZkStore) ListOfferallocatorItems() []*OfferAllocatorItem {
	return nil
}
