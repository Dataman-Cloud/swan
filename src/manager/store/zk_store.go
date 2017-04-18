package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	zookeeper "github.com/samuel/go-zookeeper/zk"
)

var (
	ZK_DEFAULT_ACL = zookeeper.WorldACL(zookeeper.PermAll)
)

const (
	OP_ADD = iota << 1
	OP_REMOVE
	OP_UPDATE
)

const (
	ENTITY_APP = iota
	ENTITY_SLOT
	ENTITY_VERSION
	ENTITY_CURRENT_TASK
	ENTITY_FRAMEWORKID
	ENTITY_OFFER_ALLOCATOR_ITEM
)

var (
	ErrAppNotFound          = errors.New("app not found")
	ErrAppAlreadyExists     = errors.New("app already exists")
	ErrSlotNotFound         = errors.New("slot not found")
	ErrSlotAlreadyExists    = errors.New("slot already exists")
	ErrVersionAlreadyExists = errors.New("version already exists")
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
	Slots    map[string]*slotStorage
}

type ZkStore struct {
	Apps           map[string]*appStorage
	OfferAllocator map[string]*OfferAllocatorItem
	FrameworkId    string

	mu   sync.RWMutex
	conn *zookeeper.Conn
}

func NewZkStore() *ZkStore {
	conn, _, err := zookeeper.Connect([]string{"114.55.130.152:2181"}, 5*time.Second)
	if err != nil {
		panic(err)
	}

	return &ZkStore{
		conn:           conn,
		Apps:           make(map[string]*appStorage),
		OfferAllocator: make(map[string]*OfferAllocatorItem),
	}
}

func (zk *ZkStore) Apply(op *StoreOp) error {
	//zk.Storage.Apply(op)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(op)
	if err != nil {
		return err
	}

	fmt.Println(op.Op, "  ", op.Entity)

	switch op.Entity {
	case ENTITY_FRAMEWORKID:
		zk.applyFrameworkId(op.Op, op)
	case ENTITY_APP:
		zk.applyApp(op.Op, op)
	case ENTITY_SLOT:
		zk.applySlot(op.Op, op)
	case ENTITY_VERSION:
		zk.applyVersion(op.Op, op)
	case ENTITY_CURRENT_TASK:
		zk.applyCurrentTask(op.Op, op)
	case ENTITY_OFFER_ALLOCATOR_ITEM:
		zk.applyOfferAllocatorItem(op.Op, op)
	default:
		panic("invalid entity type")
	}

	//foo, err := zk.conn.Create("/swan/store-op/fsysy", buf.Bytes(), zookeeper.FlagSequence, ZK_DEFAULT_ACL)
	//fmt.Println("foooooooooooooooooooooooooooooo")
	//fmt.Println(foo)
	return nil
}

func (zk *ZkStore) applyOfferAllocatorItem(operation uint8, op *StoreOp) {
	switch operation {
	case OP_ADD:
		zk.OfferAllocator[op.Param1] = op.Payload.(*OfferAllocatorItem)
	case OP_REMOVE:
		delete(zk.OfferAllocator, op.Param1)
	case OP_UPDATE:
		delete(zk.OfferAllocator, op.Param1)
		zk.OfferAllocator[op.Param1] = op.Payload.(*OfferAllocatorItem)
	default:
		panic("applyFrameworkId not supportted operation")
	}
}

func (zk *ZkStore) applyCurrentTask(operation uint8, op *StoreOp) {
	switch operation {
	case OP_UPDATE:
		zk.Apps[op.Param1].Slots[op.Param2].CurrentTask = op.Payload.(*Task)
	default:
		panic("applySlot not supportted operation")
	}
}

func (zk *ZkStore) applySlot(operation uint8, op *StoreOp) {
	switch operation {
	case OP_ADD:
		zk.Apps[op.Param1].Slots[op.Param2] = &slotStorage{Slot: op.Payload.(*Slot)}
	case OP_REMOVE:
		delete(zk.Apps[op.Param1].Slots, op.Param2)
	case OP_UPDATE:
		delete(zk.Apps[op.Param1].Slots, op.Param2)
		zk.Apps[op.Param1].Slots[op.Param2] = &slotStorage{Slot: op.Payload.(*Slot)}
	default:
		panic("applySlot not supportted operation")
	}
}

func (zk *ZkStore) applyVersion(operation uint8, op *StoreOp) {
	switch operation {
	case OP_ADD:
		zk.Apps[op.Param1].Versions[op.Param2] = op.Payload.(*Version)
	default:
		panic("applyVersion not supportted operation")
	}
}

func (zk *ZkStore) applyFrameworkId(operation uint8, op *StoreOp) {
	switch operation {
	case OP_ADD:
		zk.FrameworkId = op.Payload.(string)
	case OP_REMOVE:
		zk.FrameworkId = ""
	case OP_UPDATE:
		zk.FrameworkId = op.Payload.(string)
	default:
		panic("applyFrameworkId not supportted operation")
	}
}

func (zk *ZkStore) applyApp(operation uint8, op *StoreOp) {
	switch operation {
	case OP_ADD:
		zk.Apps[op.Param1] = &appStorage{
			App:      op.Payload.(*Application),
			Versions: make(map[string]*Version, 0),
			Slots:    make(map[string]*slotStorage),
		}
	case OP_REMOVE:
		delete(zk.Apps, op.Param1)

	case OP_UPDATE:
		if _, found := zk.Apps[op.Param1]; found {
			zk.Apps[op.Param1] = &appStorage{
				App:      op.Payload.(*Application),
				Versions: make(map[string]*Version, 0),
				Slots:    make(map[string]*slotStorage),
			}
		}

	default:
		panic("applyApp not supportted operation")
	}
}
