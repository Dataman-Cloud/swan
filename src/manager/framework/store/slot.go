package store

import (
	raftstore "github.com/Dataman-Cloud/swan/src/manager/raft/store"
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"
	"github.com/boltdb/bolt"

	"golang.org/x/net/context"
)

func (s *FrameworkStore) CreateSlot(ctx context.Context, slot *types.Slot, cb func()) error {
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Slot{slot},
	}}

	return s.RaftNode.ProposeValue(ctx, storeActions, cb)
}

func (s *FrameworkStore) GetSlot(appId, slotId string) (*types.Slot, error) {
	slot := &types.Slot{}

	if err := s.BoltbDb.View(func(tx *bolt.Tx) error {
		return raftstore.WithSlotBucket(tx, appId, slotId, func(bkt *bolt.Bucket) error {
			p := bkt.Get(raftstore.BucketKeyData)

			return slot.Unmarshal(p)
		})
	}); err != nil {
		return nil, err
	}

	return slot, nil
}

func (s *FrameworkStore) ListSlots(appId string) ([]*types.Slot, error) {
	var slots []*types.Slot

	if err := s.BoltbDb.View(func(tx *bolt.Tx) error {
		bkt := raftstore.GetSlotsBucket(tx, appId)
		if bkt == nil {
			slots = []*types.Slot{}
			return nil
		}

		return bkt.ForEach(func(k, v []byte) error {
			slotBucket := raftstore.GetSlotBucket(tx, appId, string(k))
			if slotBucket == nil {
				return nil
			}

			slot := &types.Slot{}
			p := slotBucket.Get(raftstore.BucketKeyData)
			if err := slot.Unmarshal(p); err != nil {
				return err
			}

			slots = append(slots, slot)
			return nil
		})

	}); err != nil {
		return nil, err
	}

	return slots, nil
}

func (s *FrameworkStore) DeleteSlot(ctx context.Context, appID, slotID string, cb func()) error {
	removeSlot := &types.Slot{AppID: appID, ID: slotID}
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_Slot{removeSlot},
	}}

	return s.RaftNode.ProposeValue(ctx, storeActions, cb)
}

func (s *FrameworkStore) UpdateSlot(ctx context.Context, slot *types.Slot, cb func()) error {
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Slot{slot},
	}}

	return s.RaftNode.ProposeValue(ctx, storeActions, cb)
}
