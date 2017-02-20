package store

import (
	"testing"

	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func listOfferallocatorItems(db *BoltbDb) ([]*types.OfferAllocatorItem, error) {
	var items []*types.OfferAllocatorItem

	if err := db.View(func(tx *bolt.Tx) error {
		bkt := GetOfferAllocatorItemBucket(tx)
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

func TestCreateOffer(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	offerAllocator := &types.OfferAllocatorItem{
		OfferID: "foo-bar",
		SlotID:  "slot1",
	}

	createStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_OfferAllocatorItem{offerAllocator},
	}}

	err := db.DoStoreActions(createStoreActions)
	assert.NoError(t, err)

	items, err := listOfferallocatorItems(db)
	assert.NoError(t, err)
	assert.NotNil(t, items)
	assert.Equal(t, len(items), 1)
	assert.Equal(t, items[0].OfferID, "foo-bar")

	cleanup()
}

func TestRemoveOffer(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	offerAllocator := &types.OfferAllocatorItem{
		OfferID: "foo-bar",
		SlotID:  "slot1",
	}

	createStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_OfferAllocatorItem{offerAllocator},
	}}

	err := db.DoStoreActions(createStoreActions)
	assert.NoError(t, err)

	items, err := listOfferallocatorItems(db)
	assert.NoError(t, err)
	assert.NotNil(t, items)
	assert.Equal(t, len(items), 1)
	assert.Equal(t, items[0].OfferID, "foo-bar")

	removeStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_OfferAllocatorItem{offerAllocator},
	}}

	err = db.DoStoreActions(removeStoreActions)
	assert.NoError(t, err)

	items, err = listOfferallocatorItems(db)
	assert.NoError(t, err)
	assert.Nil(t, items)

	cleanup()
}

func TestRmoveUnknownOffer(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	offerAllocator := &types.OfferAllocatorItem{
		OfferID: "foo-bar",
		SlotID:  "slot1",
	}

	createStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_OfferAllocatorItem{offerAllocator},
	}}

	err := db.DoStoreActions(createStoreActions)
	assert.NoError(t, err)

	items, err := listOfferallocatorItems(db)
	assert.NoError(t, err)
	assert.NotNil(t, items)
	assert.Equal(t, len(items), 1)
	assert.Equal(t, items[0].OfferID, "foo-bar")

	offerAllocator.SlotID = "slot2"
	removeStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_OfferAllocatorItem{offerAllocator},
	}}

	err = db.DoStoreActions(removeStoreActions)
	assert.NoError(t, err)

	items, err = listOfferallocatorItems(db)
	assert.NoError(t, err)

	cleanup()
}

func TestRemoveOfferWithoutCreate(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	offerAllocator := &types.OfferAllocatorItem{
		OfferID: "foo-bar",
		SlotID:  "slot1",
	}

	removeStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_OfferAllocatorItem{offerAllocator},
	}}

	err := db.DoStoreActions(removeStoreActions)
	assert.NoError(t, err)

	cleanup()
}

func TestUnknownOfferAction(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	offerAllocator := &types.OfferAllocatorItem{
		OfferID: "foo-bar",
		SlotID:  "slot1",
	}

	createStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUnknown,
		Target: &types.StoreAction_OfferAllocatorItem{offerAllocator},
	}}

	err := db.DoStoreActions(createStoreActions)
	assert.Equal(t, err, ErrUndefineOfferAllocatorItemAction)

	cleanup()
}
