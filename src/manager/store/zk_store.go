package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Dataman-Cloud/swan/src/utils"

	"github.com/Sirupsen/logrus"
	zookeeper "github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"
)

const (
	SWAN_ATOMIC_STORE_NODE_PATH = "%s/atomic-store"
	SWAN_SNAPSHOT_PATH          = "%s/snapshot"
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

type appHolder struct {
	App      *Application        `json:"app"`
	Versions map[string]*Version `json:"versions"`
	Slots    map[string]*Slot    `json:"slots"`
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

var (
	zs   *ZkStore
	once sync.Once
)

type ZkStore struct {
	Storage                  *Storage
	lastSequentialZkNodePath string
	lastSnapshotRevision     string

	readyToSnapshot bool
	mu              sync.RWMutex
	conn            *zookeeper.Conn
	zkPath          *url.URL
}

func DB() *ZkStore {
	return zs
}

func InitZkStore(zkPath *url.URL) error {
	conn, _, err := zookeeper.Connect(strings.Split(zkPath.Host, ","), 5*time.Second)
	if err != nil {
		return err
	}
	// TODO seems lack of re-connecting logic

	once.Do(func() {
		zs = &ZkStore{
			conn:            conn,
			Storage:         NewStorage(),
			zkPath:          zkPath,
			readyToSnapshot: false,
		}
	})
	return nil
}

func (zk *ZkStore) Start(ctx context.Context) error {
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			if zk.readyToSnapshot && zk.lastSnapshotRevision != zk.lastSequentialZkNodePath {
				revisionSnapshotted, err := zk.snapshot()
				if err != nil {
					logrus.Error(err)
				}

				err = zk.removeStaleAtomicOp(revisionSnapshotted)
				if err != nil {
					logrus.Error(err)
				}
			}
		case <-ctx.Done():
			zk.readyToSnapshot = false
			logrus.Info("store shutdown snapshot goroutine by ctx cancel")
			return ctx.Err()
		}
	}
}

func (zk *ZkStore) Apply(op *AtomicOp, zkPersistNeeded bool) error {
	zk.mu.Lock()
	defer zk.mu.Unlock()
	logrus.Debugf("Appling %s %s", op.Op.String(), op.Entity.String())

	var applyOk bool
	switch op.Entity {
	case ENTITY_FRAMEWORKID:
		applyOk = zk.applyFrameworkId(op)
	case ENTITY_APP:
		applyOk = zk.applyApp(op)
	case ENTITY_SLOT:
		applyOk = zk.applySlot(op)
	case ENTITY_VERSION:
		applyOk = zk.applyVersion(op)
	case ENTITY_CURRENT_TASK:
		applyOk = zk.applyCurrentTask(op)
	case ENTITY_OFFER_ALLOCATOR_ITEM:
		applyOk = zk.applyOfferAllocatorItem(op)
	default:
		panic("invalid entity type")
	}

	if zkPersistNeeded && applyOk {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		err := enc.Encode(op)
		if err != nil {
			return err
		}

		nodePath := filepath.Join(fmt.Sprintf(SWAN_ATOMIC_STORE_NODE_PATH, zk.zkPath.Path), "prefix")
		zk.lastSequentialZkNodePath, err = zk.conn.Create(
			nodePath,
			buf.Bytes(),
			zookeeper.FlagSequence,
			ZK_DEFAULT_ACL)

		if err != nil {
			return err
		}

		logrus.Debugf("create sequence node path is %s", zk.lastSequentialZkNodePath)
	}

	return nil
}

func (zk *ZkStore) applyOfferAllocatorItem(op *AtomicOp) bool {
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

	return true
}

func (zk *ZkStore) applyCurrentTask(op *AtomicOp) bool {
	_, ok := zk.Storage.Apps[op.Param1]
	if !ok {
		return false
	}

	_, ok = zk.Storage.Apps[op.Param1].Slots[op.Param2]
	if !ok {
		return false
	}

	switch op.Op {
	case OP_UPDATE:
		zk.Storage.Apps[op.Param1].Slots[op.Param2].CurrentTask = op.Payload.(*Task)
	default:
		panic("applyCurrentTask not supportted operation")
	}

	return true
}

func (zk *ZkStore) applySlot(op *AtomicOp) bool {
	_, ok := zk.Storage.Apps[op.Param1]
	if !ok {
		return false
	}

	switch op.Op {
	case OP_ADD:
		zk.Storage.Apps[op.Param1].Slots[op.Param2] = op.Payload.(*Slot)
	case OP_REMOVE:
		delete(zk.Storage.Apps[op.Param1].Slots, op.Param2)
	case OP_UPDATE:
		zk.Storage.Apps[op.Param1].Slots[op.Param2] = op.Payload.(*Slot)
	default:
		panic("applySlot not supported operation")
	}

	return true
}

func (zk *ZkStore) applyVersion(op *AtomicOp) bool {
	_, ok := zk.Storage.Apps[op.Param1]
	if !ok {
		return false
	}

	switch op.Op {
	case OP_ADD:
		zk.Storage.Apps[op.Param1].Versions[op.Param2] = op.Payload.(*Version)
	default:
		panic("applyVersion not supportted operation")
	}

	return true
}

func (zk *ZkStore) applyFrameworkId(op *AtomicOp) bool {
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

	return true
}

func (zk *ZkStore) applyApp(op *AtomicOp) bool {
	switch op.Op {
	case OP_ADD:
		zk.Storage.Apps[op.Param1] = &appHolder{
			App:      op.Payload.(*Application),
			Versions: make(map[string]*Version),
			Slots:    make(map[string]*Slot),
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
	return true
}

// remove any atomic op that are already snapshotted
func (zk *ZkStore) removeStaleAtomicOp(snapshotTo string) error {
	atomicStorePath := fmt.Sprintf(SWAN_ATOMIC_STORE_NODE_PATH, zk.zkPath.Path)
	children, _, err := zk.conn.Children(atomicStorePath)
	if err != nil {
		return err
	}

	for _, child := range children {
		if strings.Compare(snapshotTo, child) != 1 {
			logrus.Debugf("deleting %s now", child)
			err := zk.conn.Delete(filepath.Join(atomicStorePath, child), -1)
			if err != nil {
				return err
			}
		}
	}

	logrus.Debugf("revision before %s are cleared", snapshotTo)
	return nil
}

func (zk *ZkStore) snapshot() (string, error) {
	snapshotPath := fmt.Sprintf(SWAN_SNAPSHOT_PATH, zk.zkPath.Path)
	revision := zk.lastSequentialZkNodePath

	zk.mu.Lock()
	data, err := json.Marshal(zk.Storage)
	if err != nil {
		return "", err
	}
	zk.mu.Unlock()

	logrus.Debugf("snapshot storage to zk with data len: %d", len(data))
	logrus.Debugf(string(data))

	exists, _, err := zk.conn.Exists(snapshotPath)
	if err != nil {
		return "", err
	}

	if !exists {
		_, err = zk.conn.Create(snapshotPath, data, 0, ZK_DEFAULT_ACL)
	} else {
		_, err = zk.conn.Set(snapshotPath, data, -1)
	}
	if err != nil {
		return "", err
	}

	zk.lastSnapshotRevision = revision
	logrus.Debugf("snapshot %s to zk success", revision)

	return revision, nil
}

func (zk *ZkStore) Recover() error {
	if err := zk.recoverFromSnapshot(); err != nil {
		return err
	}

	if err := zk.recoverFromAtomicSequentialSlice(); err != nil {
		return err
	}

	zk.readyToSnapshot = true

	return nil
}

func (zk *ZkStore) recoverFromSnapshot() error {
	logrus.Debugf("syncFromSnapshot now")

	snapshotPath := fmt.Sprintf(SWAN_SNAPSHOT_PATH, zk.zkPath.Path)
	exists, _, err := zk.conn.Exists(snapshotPath)
	if err != nil {
		return err
	}

	if !exists { // do nothing when fresh start
		return nil
	}

	data, _, err := zk.conn.Get(snapshotPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, zk.Storage)
	if err != nil {
		return err
	}

	return nil
}

func (zk *ZkStore) recoverFromAtomicSequentialSlice() error {
	atomicStorePath := fmt.Sprintf(SWAN_ATOMIC_STORE_NODE_PATH, zk.zkPath.Path)

	children, _, err := zk.conn.Children(atomicStorePath)
	if err != nil {
		return err
	}

	sortedPaths := utils.SortableNodePath(children)
	sort.Sort(sortedPaths)

	for _, child := range sortedPaths {
		data, _, err := zk.conn.Get(filepath.Join(atomicStorePath, child))
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
