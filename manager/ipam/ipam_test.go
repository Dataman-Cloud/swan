package ipam

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIPAM(t *testing.T) {
	bolt, err := NewBoltStore("/tmp/xxxx")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/xxxx")
	}()

	assert.Nil(t, err)

	ipam := NewIPAM(bolt)
	assert.NotNil(t, ipam)
}

func TestRefillIp(t *testing.T) {
	bolt, err := NewBoltStore("/tmp/xxxx")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/xxxx")
	}()
	ipam := NewIPAM(bolt)

	ip1 := IP{Ip: "192.168.1.1"}
	ip2 := IP{Ip: "192.168.1.2"}
	list := IPList([]IP{ip1, ip2})
	err = ipam.Refill(list)
	assert.Nil(t, err)

	listRet, err := ipam.AllIPs()

	assert.Equal(t, 2, len(listRet))
}

func TestClear(t *testing.T) {
	bolt, err := NewBoltStore("/tmp/xxxx")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/xxxx")
	}()
	ipam := NewIPAM(bolt)

	ip1 := IP{Ip: "192.168.1.1"}
	ip2 := IP{Ip: "192.168.1.2"}
	list := IPList([]IP{ip1, ip2})
	err = ipam.Refill(list)
	assert.Nil(t, err)

	ipam.Clear()

	listRet, err := ipam.AllIPs()
	assert.Equal(t, 0, len(listRet))
}

func TestAllIps(t *testing.T) {
	bolt, err := NewBoltStore("/tmp/xxxx")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/xxxx")
	}()
	ipam := NewIPAM(bolt)

	ip1 := IP{Ip: "192.168.1.1"}
	ip2 := IP{Ip: "192.168.1.2"}
	list := IPList([]IP{ip1, ip2})
	err = ipam.Refill(list)
	assert.Nil(t, err)

	listRet, err := ipam.AllIPs()
	assert.Equal(t, 2, len(listRet))
}

func TestAllIpAvailable(t *testing.T) {
	bolt, err := NewBoltStore("/tmp/xxxx")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/xxxx")
	}()
	ipam := NewIPAM(bolt)

	ip1 := IP{Ip: "192.168.1.1", State: IP_STATE_AVAILABLE}
	ip2 := IP{Ip: "192.168.1.2"}
	list := IPList([]IP{ip1, ip2})
	err = ipam.Refill(list)
	assert.Nil(t, err)

	ip1.State = IP_STATE_ALLOCATED
	ipam.AllocateIp(ip1)
	listRet, err := ipam.IPsAvailable()
	assert.Equal(t, 1, len(listRet))
}

func TestAllIpsAllocated(t *testing.T) {
	bolt, err := NewBoltStore("/tmp/xxxx")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/xxxx")
	}()
	ipam := NewIPAM(bolt)

	ip1 := IP{Ip: "192.168.1.1", State: IP_STATE_AVAILABLE}
	ip2 := IP{Ip: "192.168.1.2"}
	list := IPList([]IP{ip1, ip2})
	err = ipam.Refill(list)
	assert.Nil(t, err)

	ip1.State = IP_STATE_ALLOCATED
	ipam.AllocateIp(ip1)
	listRet, err := ipam.IPsAllocated()
	assert.Equal(t, 1, len(listRet))
}

func TestAllocatedNextAvailabelIP(t *testing.T) {
	bolt, err := NewBoltStore("/tmp/xxxx")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/xxxx")
	}()
	ipam := NewIPAM(bolt)

	ip1 := IP{Ip: "192.168.1.1"}
	ip2 := IP{Ip: "192.168.1.2"}
	list := IPList([]IP{ip1, ip2})
	err = ipam.Refill(list)
	assert.Nil(t, err)

	ip, err := ipam.AllocateNextAvailableIP()
	assert.Nil(t, err)
	assert.Equal(t, "192.168.1.1", ip.Ip)

	ip1, err1 := ipam.AllocateNextAvailableIP()
	assert.Nil(t, err1)
	assert.Equal(t, "192.168.1.2", ip1.Ip)
}

func TestAllocatedIp(t *testing.T) {
	bolt, err := NewBoltStore("/tmp/xxxx")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/xxxx")
	}()
	ipam := NewIPAM(bolt)

	ip1 := IP{Ip: "192.168.1.1"}
	ip2 := IP{Ip: "192.168.1.2"}
	list := IPList([]IP{ip1, ip2})
	err = ipam.Refill(list)
	assert.Nil(t, err)

	ip, err := ipam.AllocateIp(IP{Ip: "192.168.1.1"})
	assert.Nil(t, err)
	assert.Equal(t, "192.168.1.1", ip.Ip)
}

func TestAllocatedIpEmptyPool(t *testing.T) {
	bolt, err := NewBoltStore("/tmp/xxxx")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/xxxx")
	}()
	ipam := NewIPAM(bolt)

	ip1 := IP{Ip: "192.168.1.1"}
	ip2 := IP{Ip: "192.168.1.2"}
	list := IPList([]IP{ip1, ip2})
	err = ipam.Refill(list)
	assert.Nil(t, err)

	ip, err := ipam.AllocateIp(IP{Ip: "192.168.1.1"})
	assert.Nil(t, err)
	assert.Equal(t, "192.168.1.1", ip.Ip)

	_, err1 := ipam.AllocateIp(IP{Ip: "192.168.1.1"})
	assert.NotNil(t, err1)
	assert.Equal(t, ErrIpRequestedAllocated, err1)
}

func TestAllocatedIpAlreadyAllocated(t *testing.T) {
	bolt, err := NewBoltStore("/tmp/xxxx")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/xxxx")
	}()
	ipam := NewIPAM(bolt)

	ip1 := IP{Ip: "192.168.1.1"}
	list := IPList([]IP{ip1})
	err = ipam.Refill(list)
	assert.Nil(t, err)

	ip, err := ipam.AllocateIp(IP{Ip: "192.168.1.1"})
	assert.Nil(t, err)
	assert.Equal(t, "192.168.1.1", ip.Ip)

	_, err1 := ipam.AllocateIp(IP{Ip: "192.168.1.1"})
	assert.NotNil(t, err1)
	assert.Equal(t, ErrIpRequestedAllocated, err1)
}

func TestRelease(t *testing.T) {
	bolt, err := NewBoltStore("/tmp/xxxx")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/xxxx")
	}()
	ipam := NewIPAM(bolt)

	ip1 := IP{Ip: "192.168.1.1"}
	ip2 := IP{Ip: "192.168.1.2"}
	list := IPList([]IP{ip1, ip2})
	err = ipam.Refill(list)
	assert.Nil(t, err)

	ip, err := ipam.AllocateIp(IP{Ip: "192.168.1.1"})
	assert.Nil(t, err)
	assert.Equal(t, "192.168.1.1", ip.Ip)

	err1 := ipam.Release(IP{Ip: "192.168.1.1"})
	assert.Nil(t, err1)
}
