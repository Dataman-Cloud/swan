package ipam

import (
	"net"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIpToString(t *testing.T) {
	ip := IP{Ip: "127.0.0.1", State: IP_STATE_AVAILABLE, TaskId: "task-id"}
	assert.Equal(t, ip.ToString(), "ip<127.0.0.1> state<available> taskId<task-id>")
}

func TestIpKey(t *testing.T) {
	ip := IP{Ip: "127.0.0.1", State: IP_STATE_AVAILABLE, TaskId: "task-id"}
	assert.Equal(t, "127.0.0.1", ip.Key())
}

func TestToIP(t *testing.T) {
	ip := IP{Ip: "127.0.0.1", State: IP_STATE_AVAILABLE, TaskId: "task-id"}
	ipExpected := net.ParseIP("127.0.0.1")
	assert.Equal(t, ipExpected, ip.ToIP())
}

func TestToInteger(t *testing.T) {
	ip := IP{Ip: "127.0.0.1", State: IP_STATE_AVAILABLE, TaskId: "task-id"}
	i, _ := strconv.ParseInt("01111111000000000000000000000001", 2, 64)
	assert.Equal(t, i, ip.ToInteger())
}

func TestIPListSort(t *testing.T) {
	ip1 := IP{Ip: "127.0.0.1", State: IP_STATE_AVAILABLE, TaskId: "task-id"}
	ip2 := IP{Ip: "127.0.0.2", State: IP_STATE_AVAILABLE, TaskId: "task-id"}

	ips := IPList([]IP{ip2, ip1})
	assert.Equal(t, 2, len(ips))
	assert.Equal(t, ip2, ips[0])
	sort.Sort(ips)
	assert.Equal(t, ip1, ips[0])
}

func TestValidIp(t *testing.T) {
	nilIP := "foobar"
	assert.Equal(t, ErrParseIp, validIp(nilIP))

	unspeficiedIP := "0.0.0.0"
	assert.Equal(t, ErrIsUnspecified, validIp(unspeficiedIP))

	loobackIP := "127.0.0.1"
	assert.Equal(t, ErrIsLoopbackIP, validIp(loobackIP))

	linkLocalUnicast := "169.254.1.1"
	assert.Equal(t, ErrIsLinkLocalUnicast, validIp(linkLocalUnicast))

	multicast := "224.0.0.0"
	assert.Equal(t, ErrIsMultiCastIP, validIp(multicast))

	validIpStr := "192.168.1.2"
	assert.Nil(t, validIp(validIpStr))

}
