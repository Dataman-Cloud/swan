package store

import (
	raftstore "github.com/Dataman-Cloud/swan/src/manager/raft/store"
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"
	"github.com/boltdb/bolt"

	"golang.org/x/net/context"
)

func (s *FrameworkStore) CreateOfferAllocatorItem(ctx context.Context, item *types.OfferAllocatorItem, cb func()) error {
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_OfferAllocatorItem{item},
	}}

	return s.RaftNode.ProposeValue(ctx, storeActions, cb)
}

func (s *FrameworkStore) ListOfferallocatorItems() ([]*types.OfferAllocatorItem, error) {
	var items []*types.OfferAllocatorItem

	if err := s.BoltbDb.View(func(tx *bolt.Tx) error {
		bkt := raftstore.GetOfferAllocatorItemBucket(tx)
		if bkt == nil {
			items = []*types.OfferAllocatorItem{}
			return nil
		}

		return bkt.ForEach(func(k, v []byte) error {
			item := &types.OfferAllocatorItem{}
			if err := item.Unmarshal(v); err != nil {
				return err
			}

			items = append(items, item)
			return nil
		})

	}); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *FrameworkStore) DeleteOfferAllocatorItem(ctx context.Context, slotID string, cb func()) error {
	removeOfferAllocatorItem := &types.OfferAllocatorItem{SlotID: slotID}
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_OfferAllocatorItem{removeOfferAllocatorItem},
	}}

	return s.RaftNode.ProposeValue(ctx, storeActions, cb)
}
