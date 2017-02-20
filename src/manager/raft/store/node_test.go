package store

import (
	"testing"

	"github.com/Dataman-Cloud/swan/src/manager/raft/types"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestCreateNode(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	testNode := &types.Node{
		ID:            "foo-bar",
		AdvertiseAddr: "0.0.0.0:9999",
	}

	createStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Node{testNode},
	}}

	err := db.DoStoreActions(createStoreActions)
	assert.NoError(t, err)

	nodes, err := db.GetNodes()
	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Equal(t, len(nodes), 1)
	assert.Equal(t, nodes[0].ID, "foo-bar")

	cleanup()
}

func TestUpdateNode(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	testNode := &types.Node{
		ID:            "foo-bar",
		AdvertiseAddr: "0.0.0.0:9999",
	}

	createStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Node{testNode},
	}}

	err := db.DoStoreActions(createStoreActions)
	assert.NoError(t, err)

	nodes, err := db.GetNodes()
	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Equal(t, len(nodes), 1)
	assert.Equal(t, nodes[0].ID, "foo-bar")

	testNode.AdvertiseAddr = "0.0.0.0:9997"
	updateStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Node{testNode},
	}}

	err = db.DoStoreActions(updateStoreActions)
	assert.NoError(t, err)

	nodes, err = db.GetNodes()
	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Equal(t, len(nodes), 1)
	assert.Equal(t, nodes[0].ID, "foo-bar")
	assert.Equal(t, nodes[0].AdvertiseAddr, "0.0.0.0:9997")

	cleanup()
}

func TestUpdateUnknownNode(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	testNode := &types.Node{
		ID:            "foo-bar",
		AdvertiseAddr: "0.0.0.0:9999",
	}

	createStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Node{testNode},
	}}

	err := db.DoStoreActions(createStoreActions)
	assert.NoError(t, err)

	nodes, err := db.GetNodes()
	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Equal(t, len(nodes), 1)
	assert.Equal(t, nodes[0].ID, "foo-bar")

	testNode.ID = "foo-bar2"
	testNode.AdvertiseAddr = "0.0.0.0:9997"
	updateStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Node{testNode},
	}}

	err = db.DoStoreActions(updateStoreActions)
	assert.Equal(t, err, ErrNodeUnknown)

	nodes, err = db.GetNodes()
	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Equal(t, len(nodes), 1)
	assert.Equal(t, nodes[0].ID, "foo-bar")
	assert.Equal(t, nodes[0].AdvertiseAddr, "0.0.0.0:9999")

	cleanup()
}

func TestRemoveNode(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	testNode := &types.Node{
		ID:            "foo-bar",
		AdvertiseAddr: "0.0.0.0:9999",
	}

	createStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Node{testNode},
	}}

	err := db.DoStoreActions(createStoreActions)
	assert.NoError(t, err)

	nodes, err := db.GetNodes()
	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Equal(t, len(nodes), 1)
	assert.Equal(t, nodes[0].ID, "foo-bar")

	testNode.ID = "foo-bar"
	removeStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_Node{testNode},
	}}

	err = db.DoStoreActions(removeStoreActions)
	assert.NoError(t, err)

	nodes, err = db.GetNodes()
	assert.NoError(t, err)
	assert.Nil(t, nodes)

	cleanup()
}

func TestRemoveUnknownNode(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	testNode := &types.Node{
		ID:            "foo-bar",
		AdvertiseAddr: "0.0.0.0:9999",
	}

	createStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Node{testNode},
	}}

	err := db.DoStoreActions(createStoreActions)
	assert.NoError(t, err)

	nodes, err := db.GetNodes()
	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Equal(t, len(nodes), 1)
	assert.Equal(t, nodes[0].ID, "foo-bar")

	testNode.ID = "foo-bar2"
	removeStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_Node{testNode},
	}}

	err = db.DoStoreActions(removeStoreActions)
	assert.Equal(t, err, bolt.ErrBucketNotFound)

	nodes, err = db.GetNodes()
	assert.NoError(t, err)
	assert.NotNil(t, nodes)

	cleanup()
}

func TestRemoveNodeWithoutInit(t *testing.T) {
	db, cleanup := storageTestEnv(t)

	testNode := &types.Node{
		ID:            "foo-bar",
		AdvertiseAddr: "0.0.0.0:9999",
	}

	removeStoreActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_Node{testNode},
	}}

	err := db.DoStoreActions(removeStoreActions)
	assert.NoError(t, err)

	nodes, err := db.GetNodes()
	assert.NoError(t, err)
	assert.NotNil(t, nodes)

	cleanup()
}
