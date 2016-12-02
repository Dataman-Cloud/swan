package store

import (
	raftstore "github.com/Dataman-Cloud/swan/src/manager/raft/store"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/boltdb/bolt"

	"golang.org/x/net/context"
)

func (store *ManagerStore) SaveFrameworkID(frameworkId string) error {
	framework := &types.Framework{frameworkId}

	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Framework{framework},
	}}

	return store.RaftNode.ProposeValue(context.TODO(), storeActions, nil)
}

func (store *ManagerStore) FetchFrameworkID() (string, error) {
	framework := &types.Framework{}

	if err := store.BoltbDb.View(func(tx *bolt.Tx) error {
		return raftstore.WithFrameworkBucket(tx, func(bkt *bolt.Bucket) error {
			p := bkt.Get(raftstore.BucketKeyData)

			return framework.Unmarshal(p)
		})

	}); err != nil {
		return "", err
	}

	return framework.ID, nil
}

func (store *ManagerStore) HasFrameworkID() (bool, error) {
	frameworkId, err := store.FetchFrameworkID()
	if err != nil {
		return false, err
	}

	if frameworkId == "" {
		return false, nil
	}

	return true, nil
}
