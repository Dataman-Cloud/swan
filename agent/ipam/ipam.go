package ipam

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/ipam"
)

type IPAM struct {
	Name string

	store     *kvStore
	storeTyp  string
	etcdAddrs []string
	zkAddrs   []string
}

func New(name string, storeTyp string, etcdAddrs, zkAddrs []string) *IPAM {
	return &IPAM{
		Name:      name,
		storeTyp:  storeTyp,
		etcdAddrs: etcdAddrs,
		zkAddrs:   zkAddrs,
	}
}

func (m *IPAM) Serve() error {
	store, err := storeSetup(m.storeTyp, m.etcdAddrs, m.zkAddrs)
	if err != nil {
		return err
	}
	m.store = store
	defer m.cleanup()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		for range ch {
			m.cleanup()
			os.Exit(0)
		}
	}()

	h := ipam.NewHandler(m)
	return h.ServeUnix(m.Name, 0)
}

func (m *IPAM) cleanup() {
	sockPath := fmt.Sprintf("/var/run/docker/plugins/%s.sock", m.Name)
	os.Remove(sockPath)
}

// GetCapabilities Called on `docker network create`
func (m *IPAM) GetCapabilities() (*ipam.CapabilitiesResponse, error) {
	log.Println("IPAM GetCapabilities")

	return &ipam.CapabilitiesResponse{
		RequiresMACAddress: true,
	}, nil
}

// GetDefaultAddressSpaces Called on `docker network create`
func (m *IPAM) GetDefaultAddressSpaces() (*ipam.AddressSpacesResponse, error) {
	log.Println("IPAM GetDefaultAddressSpaces")

	return &ipam.AddressSpacesResponse{
		LocalDefaultAddressSpace:  "swan-local",
		GlobalDefaultAddressSpace: "swan-global",
	}, nil
}

// RequestPool Called on `docker network create`
func (m *IPAM) RequestPool(req *ipam.RequestPoolRequest) (*ipam.RequestPoolResponse, error) {
	bs, _ := json.Marshal(req)
	log.Println("IPAM RequestPool request payload:", string(bs))

	// create kv subnet
	subnet, err := NewSubNet(req.Pool) // --subnet
	if err != nil {
		return nil, err
	}

	if err := m.store.CreateSubNet(subnet); err != nil {
		return nil, err
	}

	return &ipam.RequestPoolResponse{
		PoolID: subnet.ID, // 192.168.200.0
		Pool:   req.Pool,  // 192.168.200.1/24
		Data:   nil,
	}, nil
}

// ReleasePool Called on `docker network rm`
func (m *IPAM) ReleasePool(req *ipam.ReleasePoolRequest) error {
	bs, _ := json.Marshal(req)
	log.Println("IPAM ReleasePool request payload:", string(bs))

	// TODO
	return nil
}

// RequestAddress Called on `container start` and `network create --gateway`
func (m *IPAM) RequestAddress(req *ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {
	bs, _ := json.Marshal(req)
	log.Println("IPAM RequestAddress request payload:", string(bs))

	var (
		subnetID   = req.PoolID
		preferAddr = req.Address // prefered IP, container with fixed ip: `--ip`
		respAddr   string
		err        error
	)

	respAddr, err = m.store.RequestIP(subnetID, preferAddr)
	if err != nil {
		return nil, err
	}

	log.Println("IPAM allocated ip address:", respAddr)
	return &ipam.RequestAddressResponse{
		Address: respAddr,
	}, nil
}

func (m *IPAM) ReleaseAddress(req *ipam.ReleaseAddressRequest) error {
	bs, _ := json.Marshal(req)
	log.Println("IPAM ReleaseAddress request payload:", string(bs))

	var (
		subnetID = req.PoolID
		ipAddr   = req.Address
	)

	if err := m.store.ReleaseIP(subnetID, ipAddr); err != nil {
		return err
	}

	return nil
}
