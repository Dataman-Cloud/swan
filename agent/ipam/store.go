package ipam

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	etcdkv "github.com/docker/libkv/store/etcd"
	zkkv "github.com/docker/libkv/store/zookeeper"
)

var (
	keyNetwork = "swan/network" // baseKey, note: no leading '/' to make fit with zk store
	keyPool    = "pool"         // subnet sub key name
	keyConfig  = "config"       // subnet sub key name
)

var (
	errIPOutOfPool    = errors.New("ip address out of pool")
	errIPAllocated    = errors.New("ip address already allocated")
	errNoAvaliableIP  = errors.New("no avaliable ips")
	errIPRemoveDenied = errors.New("deny to remove assigned ip from pool")
)

func init() {
	etcdkv.Register()
	zkkv.Register()
}

type kvStore struct {
	kv store.Store
}

func (m *IPAM) StoreSetup() error {
	var (
		typ       = m.cfg.StoreType
		etcdAddrs = m.cfg.EtcdAddrs
		zkAddrs   = m.cfg.ZKAddrs

		kv  store.Store
		err error
	)

	switch typ {
	case "etcd":
		kv, err = libkv.NewStore(
			store.ETCD,
			etcdAddrs,
			&store.Config{
				ConnectionTimeout: time.Second * 10,
			})
	case "zk":
		kv, err = libkv.NewStore(
			store.ZK,
			zkAddrs,
			&store.Config{
				ConnectionTimeout: time.Second * 10,
			})
	}
	if err != nil {
		return err
	}

	if ok, _ := kv.Exists(keyNetwork); !ok {
		err := kv.Put(keyNetwork, nil, &store.WriteOptions{IsDir: true})
		if err != nil {
			log.Warnf("ensure base directory %s error: %v", keyNetwork, err)
		}
	}

	m.store = &kvStore{kv: kv}
	return nil
}

// Subnet
//
//
func (s *kvStore) ListSubNets() ([]*SubNet, error) {
	kvPairs, err := s.kv.List(s.normalize(""))
	if err != nil {
		return nil, err
	}

	ret := make([]*SubNet, len(kvPairs))
	for idx, kvPair := range kvPairs {
		var (
			subnetID = filepath.Base(kvPair.Key)
		)

		subnet, err := s.GetSubNet(subnetID)
		if err != nil {
			return nil, fmt.Errorf("subnet %s error: %v", subnetID, err)
		}

		ret[idx] = subnet
	}
	return ret, nil
}

func (s *kvStore) CreateSubNet(subnet *SubNet) error {
	bs, err := s.encode(subnet)
	if err != nil {
		return err
	}

	err = s.kv.Put(s.normalize(subnet.ID, keyConfig), bs, nil)
	if err != nil {
		return err
	}

	return s.kv.Put(s.normalize(subnet.ID, keyPool), nil, &store.WriteOptions{IsDir: true})
}

func (s *kvStore) GetSubNet(id string) (*SubNet, error) {
	kvPair, err := s.kv.Get(s.normalize(id, keyConfig))
	if err != nil {
		return nil, err
	}

	var subnet *SubNet
	if err := s.decode(kvPair.Value, &subnet); err != nil {
		return nil, err
	}

	return subnet, nil
}

func (s *kvStore) RemoveSubNet(id string) error {
	subnet, err := s.GetSubNet(id)
	if err != nil {
		if err == store.ErrKeyNotFound {
			return nil
		}
		return err
	}

	infos, err := s.ListIPs(subnet.ID)
	if err != nil {
		return err
	}

	var (
		n   int
		ips = make([]string, len(infos))
	)
	for idx, info := range infos {
		ips[idx] = info[0]
		if assigned, _ := strconv.ParseBool(info[1]); assigned {
			n++
		}
	}
	if n > 0 {
		return errIPRemoveDenied
	}

	if err = s.RemoveIPsFromPool(subnet.ID, ips); err != nil {
		return err
	}

	return s.kv.DeleteTree(s.normalize(subnet.ID))
}

// Subnet IPs
//
//
func (s *kvStore) ListIPs(subnetID string) ([][2]string, error) {
	kvPairs, err := s.kv.List(s.normalize(subnetID, keyPool))
	if err != nil {
		return nil, err
	}

	ret := make([][2]string, len(kvPairs))
	for idx, kvPair := range kvPairs {
		var (
			ipAddr   = filepath.Base(kvPair.Key)
			assigned = strconv.FormatBool(kvPair.Value[0] == '1')
		)
		ret[idx] = [2]string{ipAddr, assigned}
	}
	return ret, nil
}

func (s *kvStore) RequestIP(subnetID, preferIP string) (string, error) {
	subnet, err := s.GetSubNet(subnetID)
	if err != nil {
		return "", err
	}

	var (
		respIP   string
		assigned bool
	)

	// request next random subnet ip
	if preferIP == "" {
		respIP, err = s.getRandomIP(subnet.ID)
		if err != nil {
			return "", err
		}

		goto END
	}

	// request prefered IP
	assigned, err = s.checkIP(subnet.ID, preferIP)
	if err != nil {
		if err == store.ErrKeyNotFound {
			return "", errIPOutOfPool
		}
		return "", err
	}
	if assigned {
		return "", errIPAllocated
	}

	respIP = preferIP

END:

	if respIP == "" {
		return "", errNoAvaliableIP
	}

	err = s.markIPAssigned(subnet.ID, respIP)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%d", respIP, subnet.Mask), nil
}

func (s *kvStore) ReleaseIP(subnetID, ipAddr string) error {
	subnet, err := s.GetSubNet(subnetID)
	if err != nil {
		if err == store.ErrKeyNotFound {
			return nil
		}
		return err
	}

	return s.markIPFree(subnet.ID, ipAddr)
}

func (s *kvStore) AddIPsToPool(subnetID string, ips []string) error {
	_, err := s.GetSubNet(subnetID)
	if err != nil {
		return fmt.Errorf("subnet %s error: %v", subnetID, err)
	}

	for _, ip := range ips {

		var (
			key = s.normalize(subnetID, keyPool, ip)
			bs  = []byte{'0'}
		)

		if exists, _ := s.kv.Exists(key); exists {
			log.Warnf("ip %s already in the pool of subnet %s", ip, subnetID)
			continue
		}

		if err := s.kv.Put(key, bs, nil); err != nil {
			return err
		}
	}

	return nil
}

func (s *kvStore) RemoveIPsFromPool(subnetID string, ips []string) error {
	for _, ip := range ips {
		if err := s.kv.Delete(s.normalize(subnetID, keyPool, ip)); err != nil {
			if err != store.ErrKeyNotFound {
				return err
			}
		}
	}

	return nil
}

func (s *kvStore) checkIP(subnetID, ip string) (assigned bool, err error) {
	kvPair, err := s.kv.Get(s.normalize(subnetID, keyPool, ip))
	if err != nil {
		return
	}

	assigned = kvPair.Value[0] == '1'
	return
}

func (s *kvStore) getRandomIP(subnetID string) (string, error) {
	kvPairs, err := s.kv.List(s.normalize(subnetID, keyPool))
	if err != nil {
		return "", err
	}

	for _, kvPair := range kvPairs {
		if kvPair.Value[0] == '1' {
			continue
		}
		return filepath.Base(kvPair.Key), nil
	}

	return "", errNoAvaliableIP
}

func (s *kvStore) markIPAssigned(subnetID, ip string) error {
	return s.updateIP(subnetID, ip, true)
}

func (s *kvStore) markIPFree(subnetID, ip string) error {
	return s.updateIP(subnetID, ip, false)
}

func (s *kvStore) updateIP(subnetID, ip string, assigned bool) error {
	key := s.normalize(subnetID, keyPool, ip)
	if assigned {
		return s.kv.Put(key, []byte{'1'}, nil)
	}
	return s.kv.Put(key, []byte{'0'}, nil)
}

func (s *kvStore) normalize(keys ...string) string {
	elems := []string{keyNetwork}
	elems = append(elems, keys...)
	return path.Clean(path.Join(elems...))
}

// encode & decode is just short-hands for json Marshal/Unmarshal
func (s *kvStore) encode(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (s *kvStore) decode(bs []byte, v interface{}) error {
	return json.Unmarshal(bs, v)
}
