package store

import (
	raftstore "github.com/Dataman-Cloud/swan/src/manager/raft/store"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/boltdb/bolt"

	"golang.org/x/net/context"
)

// TODO(upccup): now this is not used now, but it will be useful after node can join to cluster
func (store *ManagerStore) UpdateManagerInfo(id, addr string) error {
	manager := &types.Manager{
		ID:   id,
		Addr: addr,
	}

	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Manager{manager},
	}}

	return store.RaftNode.ProposeValue(context.TODO(), storeActions, nil)
}

func (store *ManagerStore) GetManagerInfo(id string) (*types.Manager, error) {
	manager := &types.Manager{}

	if err := store.BoltbDb.View(func(tx *bolt.Tx) error {
		return raftstore.WithManagerBucket(tx, id, func(bkt *bolt.Bucket) error {
			p := bkt.Get(raftstore.BucketKeyData)

			return manager.Unmarshal(p)
		})
	}); err != nil {
		return nil, err
	}

	return manager, nil
}
