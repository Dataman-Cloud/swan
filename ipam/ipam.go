package ipam

import (
	"errors"
	"sync"
)

var (
	ErrIPAMPoolEmpty        = errors.New("ipam pool empty")
	ErrIpRequestedAllocated = errors.New("ip allocated")
	ErrIpRequestedNotFound  = errors.New("ip not found")
	ErrNotInAllocatedState  = errors.New("ip not in valid state")
)

type IPAM struct {
	mutex sync.Mutex // protected from multiple goroutine might access same IP at the same time
	store IPAMStore
}

// for time now, IPAM only acts as accessor of ips.
// it should be a sigleton but not necessary for now.
func NewIPAM(store IPAMStore) *IPAM {
	return &IPAM{
		store: store,
	}
}

func (ipam *IPAM) GetIp(key string) (IP, error) {
	return ipam.store.RetriveIP(key)
}

func (ipam *IPAM) Refill(listOfIps IPList) error {
	if err := ipam.Clear(); err != nil {
		return err
	}

	for _, ip := range listOfIps {
		ip.State = IP_STATE_AVAILABLE
		err := ipam.store.SaveIP(ip)
		if err != nil {
			return err
		}
	}

	return nil
}

// remove all IPs in the pool, used when want refresh IPAM pool
func (ipam *IPAM) Clear() error {
	return ipam.store.EmptyPool()
}

// list all ips managed by IPAM
func (ipam *IPAM) AllIPs() (IPList, error) {
	return ipam.store.ListAllIPs()
}

// list all IPs that in state `available'
func (ipam *IPAM) IPsAvailable() (IPList, error) {
	list, err := ipam.AllIPs()
	ips := []IP{}
	if err != nil {
		return IPList(ips), nil
	}

	for _, ip := range list {
		if ip.State == IP_STATE_AVAILABLE {
			ips = append(ips, ip)
		}
	}

	return IPList(ips), nil
}

// list all IPs that in state `allocated'
func (ipam *IPAM) IPsAllocated() (IPList, error) {
	list, err := ipam.AllIPs()
	ips := []IP{}
	if err != nil {
		return IPList(ips), nil
	}

	for _, ip := range list {
		if ip.State == IP_STATE_ALLOCATED {
			ips = append(ips, ip)
		}
	}

	return IPList(ips), nil
}

// retrive next avaliable IP, mark that IP as `allocated`
func (ipam *IPAM) AllocateNextAvailableIP() (IP, error) {
	iplist, err := ipam.IPsAvailable()
	if err != nil {
		return IP{}, err
	}

	if len(iplist) == 0 {
		return IP{}, ErrIPAMPoolEmpty
	}

	ipam.mutex.Lock()
	defer ipam.mutex.Unlock()

	ip := iplist[0]
	ip.State = IP_STATE_ALLOCATED
	err = ipam.store.UpdateIP(ip)
	if err != nil {
		return IP{}, err
	}

	return ip, nil
}

func (ipam *IPAM) AllocateIp(ip IP) (IP, error) {
	iplist, err := ipam.AllIPs()
	if err != nil {
		return IP{}, err
	}

	if len(iplist) == 0 {
		return IP{}, ErrIPAMPoolEmpty
	}

	ipam.mutex.Lock()
	defer ipam.mutex.Unlock()

	ipRet := IP{}
	for _, v := range iplist {
		if v.Ip == ip.Ip {
			ipRet = v
		}
	}

	if ipRet.State == IP_STATE_ALLOCATED {
		return ipRet, ErrIpRequestedAllocated
	}

	if ipRet.Ip != ip.Ip {
		return ipRet, ErrIpRequestedNotFound
	}

	ip.State = IP_STATE_ALLOCATED
	err = ipam.store.UpdateIP(ip)
	if err != nil {
		return IP{}, err
	}

	return ip, nil
}

func (ipam *IPAM) Release(ip IP) error {
	ip, err := ipam.GetIp(ip.Key())
	if err != nil {
		return err
	}

	if ip.State != IP_STATE_ALLOCATED {
		return ErrNotInAllocatedState
	}

	ip.State = IP_STATE_AVAILABLE
	err = ipam.store.UpdateIP(ip)
	if err != nil {
		return err
	}

	return nil
}
