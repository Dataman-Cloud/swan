package boltdb

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveFrameworkID(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	frameworId := "xxx-yyy-zzz"

	bolt.SaveFrameworkID(frameworId)

	Id, _ := bolt.FetchFrameworkID()

	assert.Equal(t, Id, frameworId)
}

func TestFetchFrameworkID(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	fId, _ := bolt.FetchFrameworkID()
	assert.Equal(t, fId, "")
}

func TestHasFrameworkID(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	exists, _ := bolt.HasFrameworkID()
	assert.False(t, exists)

	frameworId := "xxx-yyy-zzz"
	bolt.SaveFrameworkID(frameworId)

	exists, _ = bolt.HasFrameworkID()
	assert.True(t, exists)
}
