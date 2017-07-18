package ipam

import (
	"encoding/binary"
	"fmt"
	"net"
)

type SubNet struct {
	ID      string `json:"id"`
	CIDR    string `json:"cidr"`
	IPNet   string `json:"ipnet"`
	IPStart string `json:"ip_start"`
	IPEnd   string `json:"ip_end"`
	Mask    int    `json:"mask"`
}

func NewSubNet(cidr string) (*SubNet, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("parse CIDR %s error: %v", cidr, err)
	}

	var (
		start, end = networkIPRange(ipnet)
		ones, _    = ipnet.Mask.Size()
	)
	return &SubNet{
		ID:      ip.Mask(ipnet.Mask).String(),
		CIDR:    cidr,
		IPNet:   ipnet.String(),
		IPStart: start.String(),
		IPEnd:   end.String(),
		Mask:    ones,
	}, nil
}

func networkIPRangeList(start, end net.IP) ([]net.IP, error) {
	if start == nil {
		return nil, nil
	}

	if end == nil {
		return []net.IP{start}, nil
	}

	intStart := binary.BigEndian.Uint32(start.To4())
	intEnd := binary.BigEndian.Uint32(end.To4())
	if intStart > intEnd {
		return nil, fmt.Errorf("ip-start must be less than ip-end")
	}

	var ret []net.IP
	for i := intStart; i <= intEnd; i++ {
		ip := make([]byte, 4)
		binary.BigEndian.PutUint32(ip, i)
		ret = append(ret, net.IP(ip))
	}
	return ret, nil
}

func networkIPRange(ipnet *net.IPNet) (net.IP, net.IP) {
	if ipnet == nil {
		return nil, nil
	}

	var (
		firstIP = ipnet.IP.Mask(ipnet.Mask)
		lastIP  = cloneIP(firstIP)
	)
	for i := 0; i < len(firstIP); i++ {
		lastIP[i] = firstIP[i] | ^ipnet.Mask[i]
	}

	if ipnet.IP.To4() != nil {
		firstIP = firstIP.To4()
		lastIP = lastIP.To4()
	}

	return firstIP, lastIP
}

func cloneIP(from net.IP) net.IP {
	if from == nil {
		return nil
	}
	to := make(net.IP, len(from))
	copy(to, from)
	return to
}
