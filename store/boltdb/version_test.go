package boltdb

import (
	"github.com/Dataman-Cloud/swan/types"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSaveVersion(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	version := &types.Version{
		ID: "xxxxxx",
	}

	bolt.SaveVersion(version)

	versions, _ := bolt.ListVersions("xxxxxx")
	assert.Equal(t, len(versions), 1)
}

func TestFetchVersion(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	version := &types.Version{
		ID: "xxxxxx",
	}

	bolt.SaveVersion(version)

	versions, _ := bolt.ListVersions("xxxxxx")

	version, _ = bolt.FetchVersion(versions[0])
	assert.Equal(t, version.ID, "xxxxxx")

	version, _ = bolt.FetchVersion("yyxxzz")
	assert.Nil(t, version)
}

func TestDeleteVersion(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	version := &types.Version{
		ID: "xxxxxx",
	}

	bolt.SaveVersion(version)

	versions, _ := bolt.ListVersions("xxxxxx")

	bolt.DeleteVersion(versions[0])

	versions, _ = bolt.ListVersions("xxxxxx")
	assert.Equal(t, len(versions), 0)
}
