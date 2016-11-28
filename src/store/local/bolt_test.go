package boltdb

import (
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestNewBoltStore(t *testing.T) {
	db, err := bolt.Open("/tmp/xxxx", 0600, nil)
	assert.Nil(t, err)

	_, err = NewBoltStore(db)
	assert.Nil(t, err)

	defer func() {
		db.Close()
		os.Remove("/tmp/xxxx")
	}()
}
