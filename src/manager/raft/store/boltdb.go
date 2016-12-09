package store

import (
	"errors"

	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
)

type BoltbDb struct {
	*bolt.DB
}

var (
	bucketKeyStorageVersion = []byte("v1")
	bucketKeyApps           = []byte("apps")
	bucketKeyFramework      = []byte("framework")
	bucketKeyTasks          = []byte("tasks")
	bucketKeyVersions       = []byte("versions")

	BucketKeyData = []byte("data")
)

var (
	ErrAppUnknown              = errors.New("boltdb: app unknown")
	ErrTaskUnknown             = errors.New("boltdb: task unknown")
	ErrVersionUnknown          = errors.New("boltdb: version unknown")
	ErrNilStoreAction          = errors.New("boltdb: nil store action")
	ErrUndefineStoreAction     = errors.New("boltdb: undefined store action")
	ErrUndefineAppStoreAction  = errors.New("boltdb: undefined app store action")
	ErrUndefineFrameworkAction = errors.New("boltdb: undefined framework store action")
	ErrUndefineTaskAction      = errors.New("boltdb: undefined task store action")
	ErrUndefineVersionAction   = errors.New("boltdb: undefined version store action")
)

func NewBoltbdStore(db *bolt.DB) (*BoltbDb, error) {
	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyApps); err != nil {
			return err
		}

		return nil

	}); err != nil {
		return nil, err
	}

	return &BoltbDb{db}, nil
}

func createBucketIfNotExists(tx *bolt.Tx, keys ...[]byte) (*bolt.Bucket, error) {
	bkt, err := tx.CreateBucketIfNotExists(keys[0])
	if err != nil {
		return nil, err
	}

	for _, key := range keys[1:] {
		bkt, err = bkt.CreateBucketIfNotExists(key)
		if err != nil {
			return nil, err
		}
	}

	return bkt, nil
}

func getBucket(tx *bolt.Tx, keys ...[]byte) *bolt.Bucket {
	bkt := tx.Bucket(keys[0])

	for _, key := range keys[1:] {
		if bkt == nil {
			break
		}

		bkt = bkt.Bucket(key)
	}

	return bkt
}

func (db *BoltbDb) DoStoreActions(actions []*types.StoreAction) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, storeAction := range actions {
		if err := doStoreAction(tx, storeAction); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func doStoreAction(tx *bolt.Tx, action *types.StoreAction) error {
	if action == nil {
		return ErrNilStoreAction
	}

	actionTarget := action.GetTarget()
	if actionTarget == nil {
		return ErrUndefineStoreAction
	}

	switch actionTarget.(type) {
	case *types.StoreAction_Application:
		return doAppStoreAction(tx, action.Action, action.GetApplication())
	case *types.StoreAction_Framework:
		return doFrameworkStoreAction(tx, action.Action, action.GetFramework())
		//	case *types.StoreAction_Task:
		//		return doTaskStoreAction(tx, action.Action, action.GetTask())
		//	case *types.StoreAction_Version:
		//		return doVersionStoreAction(tx, action.Action, action.GetVersion())
	default:
		return ErrUndefineStoreAction
	}
}

func doAppStoreAction(tx *bolt.Tx, action types.StoreActionKind, app *types.Application) error {
	switch action {
	case types.StoreActionKindCreate, types.StoreActionKindUpdate:
		return putApp(tx, app)
	case types.StoreActionKindRemove:
		return removeApp(tx, app.ID)
	default:
		return ErrUndefineAppStoreAction
	}
}

func doFrameworkStoreAction(tx *bolt.Tx, action types.StoreActionKind, framework *types.Framework) error {
	switch action {
	case types.StoreActionKindCreate, types.StoreActionKindUpdate:
		return putFramework(tx, framework)
	case types.StoreActionKindRemove:
		return removeFramework(tx)
	default:
		return ErrUndefineFrameworkAction
	}
}

//func doTaskStoreAction(tx *bolt.Tx, action types.StoreActionKind, task *types.Task) error {
//	switch action {
//	case types.StoreActionKindCreate, types.StoreActionKindUpdate:
//		return putTask(tx, task)
//	case types.StoreActionKindRemove:
//		return removeTask(tx, task.AppId, task.Name)
//	default:
//		return ErrUndefineTaskAction
//	}
//}
//
//func doVersionStoreAction(tx *bolt.Tx, action types.StoreActionKind, version *types.Version) error {
//	switch action {
//	case types.StoreActionKindCreate, types.StoreActionKindUpdate:
//		return putVersion(tx, version)
//	case types.StoreActionKindRemove:
//		return removeVersion(tx, version.AppId, version.ID)
//	default:
//		return ErrUndefineVersionAction
//	}
//}
