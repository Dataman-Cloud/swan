package store

import (
	raftstore "github.com/Dataman-Cloud/swan/src/manager/raft/store"
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
	"golang.org/x/net/context"
)

func (s *FrameworkStore) UpdateFrameworkId(ctx context.Context, frameworkId string, cb func()) error {
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Framework{&types.Framework{frameworkId}},
	}}

	return s.RaftNode.ProposeValue(ctx, storeActions, cb)
}

func (s *FrameworkStore) GetFrameworkId() (string, error) {
	framework := &types.Framework{}

	if err := s.BoltbDb.View(func(tx *bolt.Tx) error {
		return raftstore.WithFrameworkBucket(tx, func(bkt *bolt.Bucket) error {
			p := bkt.Get(raftstore.BucketKeyData)

			return framework.Unmarshal(p)
		})
	}); err != nil {
		return "", err
	}

	return framework.ID, nil
}
