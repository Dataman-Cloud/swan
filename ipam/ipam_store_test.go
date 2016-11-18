package ipam

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveIP(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	ip := IP{Ip: "127.0.0.1"}

	err := bolt.SaveIP(ip)
	assert.Nil(t, err)

	ip, err1 := bolt.RetriveIP("127.0.0.1")
	assert.Nil(t, err1)

	assert.Equal(t, "127.0.0.1", ip.Ip)
}

func TestRetriveIP(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	ip := IP{Ip: "127.0.0.1"}

	err := bolt.SaveIP(ip)
	assert.Nil(t, err)

	ip, err1 := bolt.RetriveIP("127.0.0.1")
	assert.Nil(t, err1)

	assert.Equal(t, "127.0.0.1", ip.Ip)
}

func TestListIPs(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	ip := IP{Ip: "127.0.0.1"}
	err := bolt.SaveIP(ip)
	assert.Nil(t, err)

	ip1 := IP{Ip: "127.0.0.2"}
	err1 := bolt.SaveIP(ip1)
	assert.Nil(t, err1)

	list, err2 := bolt.ListAllIPs()
	assert.Nil(t, err2)
	assert.Equal(t, 2, len(list))
}

func TestUpdateIP(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	ip := IP{Ip: "127.0.0.1"}
	err := bolt.SaveIP(ip)
	assert.Nil(t, err)

	ip2 := IP{Ip: "127.0.0.1", State: IP_STATE_ALLOCATED}
	err2 := bolt.UpdateIP(ip2)
	assert.Nil(t, err2)

	ip, err1 := bolt.RetriveIP("127.0.0.1")
	assert.Nil(t, err1)

	assert.Equal(t, IP_STATE_ALLOCATED, ip.State)
}

func TestEmptyPool(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	ip := IP{Ip: "127.0.0.1"}
	err := bolt.SaveIP(ip)
	assert.Nil(t, err)

	errEmpty := bolt.EmptyPool()
	assert.Nil(t, errEmpty)

	ip, err1 := bolt.RetriveIP("127.0.0.1")
	assert.NotNil(t, err1)
	assert.Equal(t, "", ip.Ip)
}
