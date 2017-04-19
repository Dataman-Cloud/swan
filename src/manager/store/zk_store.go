package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	zookeeper "github.com/samuel/go-zookeeper/zk"
)

var (
	ZK_DEFAULT_ACL = zookeeper.WorldACL(zookeeper.PermAll)
)

// represents atomic operation both apply to ZK and intertal
// storage structure
type StoreOp uint8

var (
	OP_ADD    StoreOp = 1
	OP_REMOVE StoreOp = 2
	OP_UPDATE StoreOp = 3
)

func (op StoreOp) String() string {
	switch op {
	case OP_ADD:
		return "OP_ADD"
	case OP_REMOVE:
		return "OP_REMOVE"
	case OP_UPDATE:
		return "OP_UPDATE"
	}

	return ""
}

// represents the object type been manipulation by
// any specfic operation
type StoreEntity uint8

var (
	ENTITY_APP                  StoreEntity = 1
	ENTITY_SLOT                 StoreEntity = 2
	ENTITY_VERSION              StoreEntity = 3
	ENTITY_CURRENT_TASK         StoreEntity = 4
	ENTITY_FRAMEWORKID          StoreEntity = 5
	ENTITY_OFFER_ALLOCATOR_ITEM StoreEntity = 6
)

func (entity StoreEntity) String() string {
	switch entity {
	case ENTITY_APP:
		return "ENTITY_APP"
	case ENTITY_SLOT:
		return "ENTITY_SLOT"
	case ENTITY_VERSION:
		return "ENTITY_VERSION"
	case ENTITY_CURRENT_TASK:
		return "ENTITY_CURRENT_TASK"
	case ENTITY_FRAMEWORKID:
		return "ENTITY_FRAMEWORKID"
	case ENTITY_OFFER_ALLOCATOR_ITEM:
		return "ENTITY_OFFER_ALLOCATOR_ITEM"
	}

	return ""
}

var (
	ErrAppNotFound          = errors.New("app not found")
	ErrAppAlreadyExists     = errors.New("app already exists")
	ErrSlotNotFound         = errors.New("slot not found")
	ErrSlotAlreadyExists    = errors.New("slot already exists")
	ErrVersionAlreadyExists = errors.New("version already exists")
)

type AtomicOp struct {
	// atomic operaiton type, ADD | REMOVE | UPDATE
	Op StoreOp
	// which object type been operating
	Entity StoreEntity
	// can be explained by any specfic operation & object type operating on  appId/slotId maybe
	Param1 string
	// same as Param1
	Param2 string
	// same as Param1
	Param3 string
	// contains the data that the operation care, mostly object itself like App/Slot/Version
	Payload interface{}
}

type ZkStore struct {
	Apps           map[string]*appHolder
	OfferAllocator map[string]*OfferAllocatorItem
	FrameworkId    string

	lastSequentialZkNodePath string

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
		Apps:           make(map[string]*appHolder),
		OfferAllocator: make(map[string]*OfferAllocatorItem),
	}
}

func (zk *ZkStore) Apply(op *AtomicOp) error {
	//zk.Storage.Apply(op)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(op)
	if err != nil {
		return err
	}

	zk.mu.Lock()
	defer zk.mu.Unlock()
	switch op.Entity {
	case ENTITY_FRAMEWORKID:
		zk.applyFrameworkId(op)
	case ENTITY_APP:
		zk.applyApp(op)
	case ENTITY_SLOT:
		zk.applySlot(op)
	case ENTITY_VERSION:
		zk.applyVersion(op)
	case ENTITY_CURRENT_TASK:
		zk.applyCurrentTask(op)
	case ENTITY_OFFER_ALLOCATOR_ITEM:
		zk.applyOfferAllocatorItem(op)
	default:
		panic("invalid entity type")
	}

	zk.lastSequentialZkNodePath, err = zk.conn.Create("/swan/atomic-store/prefix",
		buf.Bytes(),
		zookeeper.FlagSequence,
		ZK_DEFAULT_ACL)
	if err != nil {
		return err
	}

	logrus.Debugf("create sequence node path is %s", zk.lastSequentialZkNodePath)
	return nil
}

func (zk *ZkStore) applyOfferAllocatorItem(op *AtomicOp) {
	switch op.Op {
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

func (zk *ZkStore) applyCurrentTask(op *AtomicOp) {
	switch op.Op {
	case OP_UPDATE:
		zk.Apps[op.Param1].Slots[op.Param2].CurrentTask = op.Payload.(*Task)
	default:
		panic("applySlot not supportted operation")
	}
}

func (zk *ZkStore) applySlot(op *AtomicOp) {
	switch op.Op {
	case OP_ADD:
		zk.Apps[op.Param1].Slots[op.Param2] = &slotHolder{Slot: op.Payload.(*Slot)}
	case OP_REMOVE:
		delete(zk.Apps[op.Param1].Slots, op.Param2)
	case OP_UPDATE:
		delete(zk.Apps[op.Param1].Slots, op.Param2)
		zk.Apps[op.Param1].Slots[op.Param2] = &slotHolder{Slot: op.Payload.(*Slot)}
	default:
		panic("applySlot not supportted operation")
	}
}

func (zk *ZkStore) applyVersion(op *AtomicOp) {
	switch op.Op {
	case OP_ADD:
		zk.Apps[op.Param1].Versions[op.Param2] = op.Payload.(*Version)
	default:
		panic("applyVersion not supportted operation")
	}
}

func (zk *ZkStore) applyFrameworkId(op *AtomicOp) {
	switch op.Op {
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

func (zk *ZkStore) applyApp(op *AtomicOp) {
	switch op.Op {
	case OP_ADD:
		zk.Apps[op.Param1] = &appHolder{
			App:      op.Payload.(*Application),
			Versions: make(map[string]*Version, 0),
			Slots:    make(map[string]*slotHolder),
		}
	case OP_REMOVE:
		delete(zk.Apps, op.Param1)

	case OP_UPDATE:
		if _, found := zk.Apps[op.Param1]; found {
			zk.Apps[op.Param1] = &appHolder{
				App:      op.Payload.(*Application),
				Versions: make(map[string]*Version, 0),
				Slots:    make(map[string]*slotHolder),
			}
		}

	default:
		panic("applyApp not supportted operation")
	}
}
