package store

import (
	"testing"

	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func getFrameworkID(db *BoltbDb) (string, error) {
	framework := &types.Framework{}

	if err := db.View(func(tx *bolt.Tx) error {
		return WithFrameworkBucket(tx, func(bkt *bolt.Bucket) error {
			p := bkt.Get(BucketKeyData)

			return framework.Unmarshal(p)
		})
	}); err != nil {
		return "", err
	}

	return framework.ID, nil
}

func TestCreateFramework(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Framework{&types.Framework{"foo-bar"}},
	}}

	err := db.DoStoreActions(storeActions)
	assert.NoError(t, err)

	frameworkID, err := getFrameworkID(db)
	assert.NoError(t, err)
	assert.Equal(t, frameworkID, "foo-bar")

	cleanup()
}

func TestRemoveFramework(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Framework{&types.Framework{"foo-bar"}},
	}}

	err := db.DoStoreActions(storeActions)
	assert.NoError(t, err)

	frameworkID, err := getFrameworkID(db)
	assert.NoError(t, err)
	assert.Equal(t, frameworkID, "foo-bar")

	removeStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_Framework{&types.Framework{"foo-bar"}},
	}}

	err = db.DoStoreActions(removeStoreActions)
	assert.NoError(t, err)

	cleanup()
}

func TestRemoveFrameworkUnknown(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	frameworkID, err := getFrameworkID(db)
	assert.NoError(t, err)
	assert.Equal(t, frameworkID, "")

	removeStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_Framework{&types.Framework{"foo-bar"}},
	}}

	err = db.DoStoreActions(removeStoreActions)
	assert.NoError(t, err)

	cleanup()
}
