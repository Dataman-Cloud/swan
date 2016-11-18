package ipam

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	IP_STATE_AVAILABLE = "available"
	IP_STATE_ALLOCATED = "allocated"
)

var (
	ErrParseIp            = errors.New("ip parse IP")
	ErrIsMultiCastIP      = errors.New("ip is a multicast IP")
	ErrIsLoopbackIP       = errors.New("ip is a loopback IP")
	ErrIsUnspecified      = errors.New("ip is unspecified")
	ErrIsLinkLocalUnicast = errors.New("ip is link local unicast")
)

// see `https://golang.org/src/net/ip.go`

type IP struct {
	Ip        string    `json:"Ip"`
	State     string    `json:"State"`
	ReleaseAt time.Time `json:"releaseAt"`
	TaskId    string    `json:"TaskId"`
}

func NewIpFromIp(ip string) (IP, error) {
	if err := validIp(ip); err != nil {
		return IP{}, err
	}

	return IP{
		Ip:    ip,
		State: IP_STATE_AVAILABLE,
	}, nil
}

func validIp(ipString string) error {
	ip := net.ParseIP(ipString)

	if ip == nil {
		return ErrParseIp
	}

	if ip.IsUnspecified() {
		return ErrIsUnspecified
	}

	if ip.IsLoopback() {
		return ErrIsLoopbackIP
	}

	if ip.IsLinkLocalUnicast() {
		return ErrIsLinkLocalUnicast
	}

	if ip.IsMulticast() {
		return ErrIsMultiCastIP
	}

	return nil
}

func (ip IP) ToIP() net.IP {
	ipv4 := net.ParseIP(ip.Ip)
	return ipv4
}

func (ip IP) ToString() string {
	return fmt.Sprintf("ip<%s> state<%s> taskId<%s>", ip.Ip, ip.State, ip.TaskId)
}

func (ip IP) ToInteger() int64 {
	ip4 := ip.ToIP().To4()
	bin := make([]string, len(ip4))
	for i, v := range ip4 {
		bin[i] = fmt.Sprintf("%08b", v)
	}
	i, _ := strconv.ParseInt(strings.Join(bin, ""), 2, 64)

	return i
}

func (ip IP) Key() string {
	return ip.Ip
}

type IPList []IP

func (s IPList) Len() int {
	return len(s)
}
func (s IPList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s IPList) Less(i, j int) bool {
	return s[i].ToInteger() < s[j].ToInteger()
}
