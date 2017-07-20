package ipam

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Dataman-Cloud/swan/config"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/ipam"
	"github.com/docker/libnetwork/netlabel"
)

type IPAM struct {
	cfg   *config.IPAM
	store *kvStore
}

func New(cfg *config.IPAM) *IPAM {
	return &IPAM{
		cfg: cfg,
	}
}

func (m *IPAM) Serve() error {
	if err := m.StoreSetup(); err != nil {
		return err
	}
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
	return h.ServeUnix("swan", 0)
}

func (m *IPAM) cleanup() {
	os.Remove("/var/run/docker/plugins/swan.sock")
}

func (m *IPAM) SetIPPool(pool *IPPoolRange) error {
	subnetID, _ := pool.SubNetID()
	return m.store.AddIPsToPool(subnetID, pool.IPList())
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

	// new subnet
	subnet, err := NewSubNet(req.Pool) // --subnet=CIDR
	if err != nil {
		log.Errorln("IPAM RequestPool NewSubNet() error: ", req.Pool, err)
		return nil, err
	}

	// check if exists
	if exists, _ := m.store.GetSubNet(subnet.ID); exists != nil {
		log.Errorln("IPAM RequestPool Conflict on: ", subnet.ID)
		return nil, errors.New("subnet already exists: " + subnet.ID)
	}

	// create kv subnet
	if err = m.store.CreateSubNet(subnet); err != nil {
		log.Errorln("IPAM RequestPool CreateSubNet() error: ", subnet.ID, err)
		return nil, err
	}

	log.Println("IPAM RequestPool succeed", req.Pool)
	return &ipam.RequestPoolResponse{
		PoolID: subnet.ID, // 192.168.200.0
		Pool:   req.Pool,  // 192.168.200.1/24 --subnet=CIDR
		Data:   nil,
	}, nil
}

// ReleasePool Called on `docker network rm`
func (m *IPAM) ReleasePool(req *ipam.ReleasePoolRequest) error {
	bs, _ := json.Marshal(req)
	log.Println("IPAM ReleasePool request payload:", string(bs))

	var (
		subnetID = req.PoolID
	)

	err := m.store.RemoveSubNet(subnetID)
	if err != nil {
		log.Errorln("IPAM ReleasePool error: ", subnetID, err)
		return err
	}

	log.Println("IPAM ReleasePool succeed", subnetID)
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

	// requeset on gateway ipaddr
	if val, ok := req.Options["RequestAddressType"]; ok && val == netlabel.Gateway {
		subnet, err := m.store.GetSubNet(subnetID)
		if err != nil {
			return nil, err
		}

		respAddr = fmt.Sprintf("%s/%d", preferAddr, subnet.Mask)
		goto END
	}

	// request on docker container start up
	respAddr, err = m.store.RequestIP(subnetID, preferAddr)
	if err != nil {
		log.Errorln("IPAM RequestAddress error:", subnetID, err)
		return nil, err
	}

END:
	log.Println("IPAM Allocated IP Address:", respAddr)
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
		log.Errorln("IPAM ReleaseAddress error: ", subnetID, ipAddr, err)
		return err
	}

	log.Println("IPAM ReleaseAddress succeed", subnetID, ipAddr)
	return nil
}
