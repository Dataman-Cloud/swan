package store

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestNewBoltdbStore(t *testing.T) {
	dir, err := ioutil.TempDir("", "testStorage")
	assert.NoError(t, err)

	dbpath := filepath.Join(dir, "swan.db")
	assert.NoError(t, os.MkdirAll(dir, 0777))
	defer os.RemoveAll(dir)

	db, err := bolt.Open(dbpath, 0666, nil)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	boltDb := NewBoltbdStore(db)
	assert.NotNil(t, boltDb)
}

// returns an initialized db and cleanup function for use in test
func storageTestEnv(t *testing.T) (*BoltbDb, func()) {
	var cleanup []func()
	dir, err := ioutil.TempDir("", "testStorage")
	assert.NoError(t, err)

	dbpath := filepath.Join(dir, "swan.db")
	assert.NoError(t, os.MkdirAll(dir, 0777))
	cleanup = append(cleanup, func() { os.RemoveAll(dir) })

	db, err := bolt.Open(dbpath, 0666, nil)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	cleanup = append(cleanup, func() { db.Close() })

	boltDb := NewBoltbdStore(db)
	assert.NotNil(t, boltDb)

	return boltDb, func() {
		//iterate in reverse so it works like defer
		for i := len(cleanup) - 1; i >= 0; i-- {
			cleanup[i]()
		}
	}
}

func TestNilAction(t *testing.T) {
	storeActions := []*types.StoreAction{nil}

	db, cleanup := storageTestEnv(t)
	err := db.DoStoreActions(storeActions)
	assert.Error(t, err)
	assert.Equal(t, err, ErrNilStoreAction)

	cleanup()

}
