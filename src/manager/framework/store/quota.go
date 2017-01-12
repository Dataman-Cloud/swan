package store

import (
	raftstore "github.com/Dataman-Cloud/swan/src/manager/raft/store"
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"
	"github.com/boltdb/bolt"

	"golang.org/x/net/context"
)

func (s *FrameworkStore) CreateQuota(ctx context.Context, resourceQuota *types.ResourceQuota, cb func()) error {
	storeAction := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_ResoureQuota{resourceQuota},
	}}

	return s.RaftNode.ProposeValue(ctx, storeAction, cb)
}

func (s *FrameworkStore) UpdateQuota(ctx context.Context, resourceQuota *types.ResourceQuota, cb func()) error {
	storeAction := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_ResoureQuota{resourceQuota},
	}}

	return s.RaftNode.ProposeValue(ctx, storeAction, cb)
}

func (s *FrameworkStore) GetQuota(quotaGroup string) (*types.ResourceQuota, error) {
	quota := &types.ResourceQuota{}

	if err := s.BoltbDb.View(func(tx *bolt.Tx) error {
		return raftstore.WithQuotaBucket(tx, quotaGroup, func(bkt *bolt.Bucket) error {
			p := bkt.Get(raftstore.BucketKeyData)

			return quota.Unmarshal(p)
		})
	}); err != nil {
		return nil, err
	}

	return quota, nil
}

func (s *FrameworkStore) ListQuotas() ([]*types.ResourceQuota, error) {
	var quotas []*types.ResourceQuota

	if err := s.BoltbDb.View(func(tx *bolt.Tx) error {
		bkt := raftstore.GetQuotasBucket(tx)
		if bkt == nil {
			quotas = []*types.ResourceQuota{}
			return nil
		}

		return bkt.ForEach(func(k, v []byte) error {
			quotaBucket := raftstore.GetQuotaBucket(tx, string(k))
			if quotaBucket == nil {
				return nil
			}

			quota := &types.ResourceQuota{}
			p := quotaBucket.Get(raftstore.BucketKeyData)
			if err := quota.Unmarshal(p); err != nil {
				return err
			}

			quotas = append(quotas, quota)
			return nil

		})

	}); err != nil {
		return nil, err
	}

	return quotas, nil
}

func (s *FrameworkStore) DeleteQuota(ctx context.Context, quotaGroup string, cb func()) error {
	removeQuota := &types.ResourceQuota{QuotaGroup: quotaGroup}
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_ResoureQuota{removeQuota},
	}}

	return s.RaftNode.ProposeValue(ctx, storeActions, cb)
}
