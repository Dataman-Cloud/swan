package ipam

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

// SubNet
//
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

// IPPoolRange
//
type IPPoolRange struct {
	IPStart string `json:"ip_start"` // CIDR
	IPEnd   string `json:"ip_end"`   // CIDR
}

func (r *IPPoolRange) SubNetID() (string, error) {
	subnet, err := NewSubNet(r.IPStart)
	if err != nil {
		return "", err
	}

	return subnet.ID, nil
}

func (r *IPPoolRange) IPList() []string {
	var (
		start, _, _ = net.ParseCIDR(r.IPStart)
		end, _, _   = net.ParseCIDR(r.IPEnd)
	)

	ret := make([]string, 0, 0)
	for _, ip := range networkIPRangeList(start, end) {
		ret = append(ret, ip.To4().String())
	}
	return ret
}

func (r *IPPoolRange) Valid() error {
	ipStart, ipNetStart, err := net.ParseCIDR(r.IPStart)
	if err != nil {
		return err
	}

	ipEnd, ipNetEnd, err := net.ParseCIDR(r.IPEnd)
	if err != nil {
		return err
	}

	if !ipNetStart.Contains(ipEnd) ||
		!ipNetEnd.Contains(ipStart) ||
		ipNetStart.Mask.String() != ipNetEnd.Mask.String() {
		return errors.New("ip-start & ip-end are not in the same subnet")
	}

	var (
		intStart = binary.BigEndian.Uint32(ipStart.To4())
		intEnd   = binary.BigEndian.Uint32(ipEnd.To4())
	)
	if intStart > intEnd {
		return errors.New("ip-start must be less than ip-end")
	}

	return nil
}

func networkIPRangeList(start, end net.IP) []net.IP {
	if start == nil {
		return nil
	}

	if end == nil {
		return []net.IP{start}
	}

	var (
		intStart = binary.BigEndian.Uint32(start.To4())
		intEnd   = binary.BigEndian.Uint32(end.To4())
	)

	var ret []net.IP
	for i := intStart; i <= intEnd; i++ {
		ip := make([]byte, 4)
		binary.BigEndian.PutUint32(ip, i)
		ret = append(ret, net.IP(ip))
	}
	return ret
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
