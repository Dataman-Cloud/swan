package store

import (
	"testing"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func getApp(db *BoltbDb, appID string) (*types.Application, error) {
	app := &types.Application{}

	if err := db.View(func(tx *bolt.Tx) error {
		return WithAppBucket(tx, appID, func(bkt *bolt.Bucket) error {
			p := bkt.Get(BucketKeyData)

			return app.Unmarshal(p)
		})
	}); err != nil {
		return nil, err
	}

	return app, nil
}

func listApps(db *BoltbDb) ([]*types.Application, error) {
	var apps []*types.Application

	if err := db.View(func(tx *bolt.Tx) error {
		bkt := GetAppsBucket(tx)
		if bkt == nil {
			apps = []*types.Application{}
			return nil
		}

		return bkt.ForEach(func(k, v []byte) error {
			appBucket := GetAppBucket(tx, string(k))
			if appBucket == nil {
				return nil
			}

			app := &types.Application{}
			p := appBucket.Get(BucketKeyData)
			if err := app.Unmarshal(p); err != nil {
				return err
			}

			apps = append(apps, app)
			return nil

		})

	}); err != nil {
		return nil, err
	}

	return apps, nil
}

func TestCreateApp(t *testing.T) {
	testApp := &types.Application{
		ID:        "foo-bar",
		Name:      "foo-bar",
		CreatedAt: time.Now().UnixNano(),
		UpdatedAt: time.Now().UnixNano(),
		State:     "normal",
	}

	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Application{testApp},
	}}

	db, cleanup := storageTestEnv(t)
	err := db.DoStoreActions(storeActions)
	assert.NoError(t, err)

	apps, err := listApps(db)
	assert.NoError(t, err)
	assert.NotNil(t, apps)
	assert.Equal(t, len(apps), 1)

	app, err := getApp(db, "foo-bar")
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, app.ID, "foo-bar")

	cleanup()
}

func TestUpdateApp(t *testing.T) {
	testApp := &types.Application{
		ID:        "foo-bar",
		Name:      "foo-bar",
		CreatedAt: time.Now().UnixNano(),
		UpdatedAt: time.Now().UnixNano(),
		State:     "normal",
	}

	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Application{testApp},
	}}

	db, cleanup := storageTestEnv(t)
	err := db.DoStoreActions(storeActions)
	assert.NoError(t, err)

	testApp.State = "pending"
	updateStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Application{testApp},
	}}

	err = db.DoStoreActions(updateStoreActions)
	assert.NoError(t, err)

	apps, err := listApps(db)
	assert.NoError(t, err)
	assert.NotNil(t, apps)
	assert.Equal(t, len(apps), 1)

	app, err := getApp(db, "foo-bar")
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, app.ID, "foo-bar")
	assert.Equal(t, app.State, "pending")

	cleanup()
}

func TestUpdateAppNotFound(t *testing.T) {
	testApp := &types.Application{
		ID:        "foo-bar",
		Name:      "foo-bar",
		CreatedAt: time.Now().UnixNano(),
		UpdatedAt: time.Now().UnixNano(),
		State:     "normal",
	}

	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Application{testApp},
	}}

	db, cleanup := storageTestEnv(t)
	err := db.DoStoreActions(storeActions)
	assert.NoError(t, err)

	testApp.ID = "foo-bar2"
	testApp.State = "pending"
	updateStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Application{testApp},
	}}

	err = db.DoStoreActions(updateStoreActions)
	assert.Error(t, err)
	assert.Equal(t, err, ErrAppUnknown)

	apps, err := listApps(db)
	assert.NoError(t, err)
	assert.NotNil(t, apps)
	assert.Equal(t, len(apps), 1)

	app, err := getApp(db, "foo-bar")
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, app.ID, "foo-bar")
	assert.Equal(t, app.State, "normal")

	cleanup()
}

func TestRemoveApp(t *testing.T) {
	testApp := &types.Application{
		ID:        "foo-bar",
		Name:      "foo-bar",
		CreatedAt: time.Now().UnixNano(),
		UpdatedAt: time.Now().UnixNano(),
		State:     "normal",
	}

	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Application{testApp},
	}}

	db, cleanup := storageTestEnv(t)
	err := db.DoStoreActions(storeActions)
	assert.NoError(t, err)

	removeStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_Application{testApp},
	}}

	err = db.DoStoreActions(removeStoreActions)
	assert.NoError(t, err)

	apps, err := listApps(db)
	assert.NoError(t, err)
	assert.Nil(t, apps)

	cleanup()
}

func TestRemoveAppNotFound(t *testing.T) {
	testApp := &types.Application{
		ID:        "foo-bar",
		Name:      "foo-bar",
		CreatedAt: time.Now().UnixNano(),
		UpdatedAt: time.Now().UnixNano(),
		State:     "normal",
	}

	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Application{testApp},
	}}

	db, cleanup := storageTestEnv(t)
	err := db.DoStoreActions(storeActions)
	assert.NoError(t, err)

	testApp.ID = "foo-bar2"
	removeStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_Application{testApp},
	}}

	err = db.DoStoreActions(removeStoreActions)
	assert.Error(t, err)
	assert.Equal(t, err, bolt.ErrBucketNotFound)

	apps, err := listApps(db)
	assert.NoError(t, err)
	assert.NotNil(t, apps)

	cleanup()
}

func TestErrorAppStoreAction(t *testing.T) {
	testApp := &types.Application{
		ID:        "foo-bar",
		Name:      "foo-bar",
		CreatedAt: time.Now().UnixNano(),
		UpdatedAt: time.Now().UnixNano(),
		State:     "normal",
	}

	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUnknown,
		Target: &types.StoreAction_Application{testApp},
	}}

	db, cleanup := storageTestEnv(t)
	err := db.DoStoreActions(storeActions)
	assert.Error(t, err)
	assert.Equal(t, err, ErrUndefineAppStoreAction)

	cleanup()
}
