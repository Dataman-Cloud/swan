package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Dataman-Cloud/swan/src/utils"

	"github.com/Sirupsen/logrus"
	zookeeper "github.com/samuel/go-zookeeper/zk"
)

const (
	ZK_SNAPSHOT_PATH = "/swan/snapshot"
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

type Storage struct {
	Apps           map[string]*appHolder          `json:"apps"`
	OfferAllocator map[string]*OfferAllocatorItem `json:"offerAllocator"`
	FrameworkId    string                         `json:"frameworkid"`
}

func NewStorage() *Storage {
	return &Storage{
		Apps:           make(map[string]*appHolder),
		OfferAllocator: make(map[string]*OfferAllocatorItem),
	}
}

type ZkStore struct {
	Storage                  *Storage
	lastSequentialZkNodePath string
	lastSnapshotRevision     string

	mu   sync.RWMutex
	conn *zookeeper.Conn
}

func NewZkStore() *ZkStore {
	conn, _, err := zookeeper.Connect([]string{"114.55.130.152:2181"}, 5*time.Second)
	if err != nil {
		panic(err)
	}
	zk := &ZkStore{
		conn:    conn,
		Storage: NewStorage(),
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				revisionSnapshotted, err := zk.snapshot()
				if err != nil {
					logrus.Error(err)
				}

				err = zk.removeStaleAtomicOp(revisionSnapshotted)
				if err != nil {
					logrus.Error(err)
				}
			}
		}
	}()

	return zk
}

func (zk *ZkStore) Apply(op *AtomicOp, zkPersistNeeded bool) error {
	zk.mu.Lock()
	defer zk.mu.Unlock()
	logrus.Debugf("Appling %s %s", op.Op.String(), op.Entity.String())

	fmt.Println("ffffffffffffffffffffffff")
	for name, app := range zk.Storage.Apps {
		fmt.Println("z000000000000000")
		fmt.Println(name)
		fmt.Println(app.Slots)
		fmt.Println(app.Versions)
	}
	x, _ := json.Marshal(zk.Storage)
	fmt.Println(string(x))

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

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(op)
	if err != nil {
		return err
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
		zk.Storage.OfferAllocator[op.Param1] = op.Payload.(*OfferAllocatorItem)
	case OP_REMOVE:
		delete(zk.Storage.OfferAllocator, op.Param1)
	case OP_UPDATE:
		delete(zk.Storage.OfferAllocator, op.Param1)
		zk.Storage.OfferAllocator[op.Param1] = op.Payload.(*OfferAllocatorItem)
	default:
		panic("applyFrameworkId not supportted operation")
	}
}

func (zk *ZkStore) applyCurrentTask(op *AtomicOp) {
	switch op.Op {
	case OP_UPDATE:
		zk.Storage.Apps[op.Param1].Slots[op.Param2].CurrentTask = op.Payload.(*Task)
	default:
		panic("applyCurrentTask not supportted operation")
	}
}

func (zk *ZkStore) applySlot(op *AtomicOp) {
	switch op.Op {
	case OP_ADD:
		zk.Storage.Apps[op.Param1].Slots[op.Param2] = &slotHolder{Slot: op.Payload.(*Slot)}
	case OP_REMOVE:
		delete(zk.Storage.Apps[op.Param1].Slots, op.Param2)
	case OP_UPDATE:
		zk.Storage.Apps[op.Param1].Slots[op.Param2].Slot = op.Payload.(*Slot)
	default:
		panic("applySlot not supported operation")
	}
}

func (zk *ZkStore) applyVersion(op *AtomicOp) {
	switch op.Op {
	case OP_ADD:
		zk.Storage.Apps[op.Param1].Versions[op.Param2] = op.Payload.(*Version)
	default:
		panic("applyVersion not supportted operation")
	}
}

func (zk *ZkStore) applyFrameworkId(op *AtomicOp) {
	switch op.Op {
	case OP_ADD:
		zk.Storage.FrameworkId = op.Payload.(string)
	case OP_REMOVE:
		zk.Storage.FrameworkId = ""
	case OP_UPDATE:
		zk.Storage.FrameworkId = op.Payload.(string)
	default:
		panic("applyFrameworkId not supportted operation")
	}
}

func (zk *ZkStore) applyApp(op *AtomicOp) {
	switch op.Op {
	case OP_ADD:
		zk.Storage.Apps[op.Param1] = &appHolder{
			App:      op.Payload.(*Application),
			Versions: make(map[string]*Version, 0),
			Slots:    make(map[string]*slotHolder),
		}
	case OP_REMOVE:
		delete(zk.Storage.Apps, op.Param1)

	case OP_UPDATE:
		if _, found := zk.Storage.Apps[op.Param1]; found {
			zk.Storage.Apps[op.Param1].App = op.Payload.(*Application)
		}

	default:
		panic("applyApp not supportted operation")
	}
}

// remove any atomic op that are already snapshotted
func (zk *ZkStore) removeStaleAtomicOp(snapshotTo string) error {
	children, _, err := zk.conn.Children("/swan/atomic-store")
	if err != nil {
		return err
	}

	for _, child := range children {
		if strings.Compare(snapshotTo, child) != 1 {
			logrus.Debugf("deleting %s now", child)
			err := zk.conn.Delete("/swan/atomic-store/"+child, -1)
			if err != nil {
				return err
			}
		}
	}

	logrus.Debugf("revision before %s are cleared", snapshotTo)
	return nil
}

func (zk *ZkStore) snapshot() (string, error) {
	zk.mu.Lock()
	revision := zk.lastSequentialZkNodePath
	data, err := json.Marshal(zk.Storage)
	if err != nil {
		return "", err
	}
	zk.mu.Unlock()
	logrus.Debugf("snapshot storage to zk with data len: %d", len(data))

	exists, _, err := zk.conn.Exists(ZK_SNAPSHOT_PATH)
	if err != nil {
		return "", err
	}

	if !exists {
		_, err = zk.conn.Create(ZK_SNAPSHOT_PATH,
			data, 0, ZK_DEFAULT_ACL)
	} else {
		_, err = zk.conn.Set(ZK_SNAPSHOT_PATH,
			data, -1)
	}
	if err != nil {
		return "", err
	}

	zk.lastSnapshotRevision = revision
	logrus.Debugf("snapshot %s to zk success", revision)

	return revision, nil
}

func (zk *ZkStore) Synchronize() error {
	if err := zk.syncFromSnapshot(); err != nil {
		return err
	}

	if err := zk.syncFromAtomicSequentialSlice(); err != nil {
		return err
	}

	return nil
}

func (zk *ZkStore) syncFromSnapshot() error {
	logrus.Debugf("syncFromSnapshot now")

	exists, _, err := zk.conn.Exists(ZK_SNAPSHOT_PATH)
	if err != nil {
		return err
	}

	if !exists { // do nothing when fresh start
		return nil
	}

	data, _, err := zk.conn.Get(ZK_SNAPSHOT_PATH)
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	err = json.Unmarshal(data, zk.Storage)
	if err != nil {
		return err
	}
	fmt.Printf("%+v", zk.Storage)

	return nil
}

func (zk *ZkStore) syncFromAtomicSequentialSlice() error {
	children, _, err := zk.conn.Children("/swan/atomic-store")
	if err != nil {
		return err
	}

	sortedPaths := utils.SortableNodePath(children)
	sort.Sort(sortedPaths)

	for _, child := range sortedPaths {
		data, _, err := zk.conn.Get("/swan/atomic-store/" + child)
		if err != nil {
			return err
		}

		ao, err := zk.unmarshalAtomicOp(data)
		if err != nil {
			return err
		}

		logrus.Debugf("decode  %s got %+v", child, ao)
		zk.Apply(ao, false)
	}

	return nil
}

func (zk *ZkStore) unmarshalAtomicOp(data []byte) (*AtomicOp, error) {
	var tmpAo struct {
		Op     StoreOp
		Entity StoreEntity
		Param1 string
		Param2 string
		Param3 string

		Payload json.RawMessage
	}

	err := json.Unmarshal(data, &tmpAo)
	if err != nil {
		return nil, err
	}

	var ao AtomicOp
	ao.Op = tmpAo.Op
	ao.Entity = tmpAo.Entity
	ao.Param1 = tmpAo.Param1
	ao.Param2 = tmpAo.Param2
	ao.Param3 = tmpAo.Param3

	if tmpAo.Payload != nil {
		switch ao.Entity {
		case ENTITY_APP:
			var app Application
			err = json.Unmarshal(tmpAo.Payload, &app)
			if err != nil {
				return nil, err
			}
			ao.Payload = &app
		case ENTITY_SLOT:
			var slot Slot
			err = json.Unmarshal(tmpAo.Payload, &slot)
			if err != nil {
				return nil, err
			}
			ao.Payload = &slot
		case ENTITY_VERSION:
			var version Version
			err = json.Unmarshal(tmpAo.Payload, &version)
			if err != nil {
				return nil, err
			}
			ao.Payload = &version
		case ENTITY_CURRENT_TASK:
			var task Task
			err = json.Unmarshal(tmpAo.Payload, &task)
			if err != nil {
				return nil, err
			}
			ao.Payload = &task

		case ENTITY_FRAMEWORKID:
			var frameworkid string
			err = json.Unmarshal(tmpAo.Payload, &frameworkid)
			if err != nil {
				return nil, err
			}
			ao.Payload = frameworkid

		case ENTITY_OFFER_ALLOCATOR_ITEM:
			var item OfferAllocatorItem
			err = json.Unmarshal(tmpAo.Payload, &item)
			if err != nil {
				return nil, err
			}
			ao.Payload = &item
		}
	}
	return &ao, nil
}
