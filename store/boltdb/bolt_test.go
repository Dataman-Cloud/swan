package boltdb

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewBoltStore(t *testing.T) {
	_, err := NewBoltStore("/xxxx/yyyy")
	assert.NotNil(t, err)

	bolt, err := NewBoltStore("/tmp/xxxx")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/xxxx")
	}()
}
