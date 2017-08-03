package etcd

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

const (
	keyApp         = "/apps"      // single app
	keyCompose     = "/composes"  // compose instance (group apps)
	keyFrameworkID = "/framework" // framework id

	keyTasks    = "tasks"    // sub key of keyApp
	keyVersions = "versions" // sub key of keyApp
)

var (
	errAppNotFound          = errors.New("app not found")
	errAppAlreadyExists     = errors.New("app already exists")
	errVersionAlreadyExists = errors.New("version already exists")
	errInstanceNotFound     = errors.New("instance not found")

	errInvalidGet  = errors.New("Get() on directory node make no sense")
	errInvalidList = errors.New("can't List() on key Node")
)

type EtcdStore struct {
	mapi   etcd.MembersAPI
	kapi   etcd.KeysAPI
	prefix string
}

func NewEtcdStore(addrs []string) (*EtcdStore, error) {
	cfg := etcd.Config{
		Endpoints:               addrs,
		Transport:               etcd.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second * 3,
	}

	etcdc, err := etcd.New(cfg)
	if err != nil {
		return nil, err
	}

	store := &EtcdStore{
		mapi:   etcd.NewMembersAPI(etcdc),
		kapi:   etcd.NewKeysAPI(etcdc),
		prefix: "/swan",
	}

	// create base keys nodes
	for _, node := range []string{keyApp, keyCompose} {
		store.ensureDir(node)
	}

	return store, nil
}

func (s *EtcdStore) clean(p string) string {
	if !strings.HasPrefix(p, s.prefix) {
		p = s.prefix + "/" + p
	}
	return path.Clean(p)
}

func (s *EtcdStore) get(key string) ([]byte, error) {
	key = s.clean(key)
	opts := &etcd.GetOptions{
		Recursive: false,
		Quorum:    true,
	}
	res, err := s.kapi.Get(context.Background(), key, opts)
	if err != nil {
		return nil, err
	}
	if res.Node.Dir {
		return nil, errInvalidGet
	}
	return []byte(res.Node.Value), nil
}

func (s *EtcdStore) list(key string) (map[string][]byte, error) {
	key = s.clean(key)
	opts := &etcd.GetOptions{
		Recursive: true,
		Quorum:    true,
		Sort:      true,
	}
	res := map[string][]byte{}
	resp, err := s.kapi.Get(context.Background(), key, opts)
	if err != nil {
		return res, err
	}
	if !resp.Node.Dir {
		return res, errInvalidList
	}
	for _, n := range resp.Node.Nodes {
		res[filepath.Base(n.Key)] = []byte(n.Value)
	}
	return res, nil
}

func (s *EtcdStore) ensureDir(key string) error {
	key = s.clean(key)
	if ok, _ := s.exists(key); ok {
		return nil
	}
	opts := &etcd.SetOptions{
		Dir: true,
	}
	_, err := s.kapi.Set(context.Background(), key, "", opts)
	if err != nil {
		log.Warnf("ensure etcd dir [%s] error: %v", key, err)
		return err
	}
	return nil
}

func (s *EtcdStore) create(key string, value []byte) error {
	return s.set(key, value, etcd.PrevNoExist)
}

func (s *EtcdStore) update(key string, value []byte) error {
	return s.set(key, value, etcd.PrevExist)
}

func (s *EtcdStore) upsert(key string, value []byte) error {
	return s.set(key, value, etcd.PrevIgnore)
}

func (s *EtcdStore) set(key string, value []byte, mode etcd.PrevExistType) error {
	key = s.clean(key)
	opts := &etcd.SetOptions{
		PrevExist: mode,
		Dir:       false,
		TTL:       time.Duration(0),
	}
	_, err := s.kapi.Set(context.Background(), key, string(value), opts)
	return err
}

func (s *EtcdStore) del(key string, recursive bool) error {
	key = s.clean(key)
	opts := &etcd.DeleteOptions{
		Dir:       false,
		Recursive: recursive,
	}
	_, err := s.kapi.Delete(context.Background(), key, opts)
	return err
}

func (s *EtcdStore) delDir(key string, recursive bool) error {
	key = s.clean(key)
	opts := &etcd.DeleteOptions{
		Dir:       true,
		Recursive: recursive,
	}
	_, err := s.kapi.Delete(context.Background(), key, opts)
	return err
}

func (s *EtcdStore) exists(key string) (bool, error) {
	_, err := s.get(key)
	if err != nil {
		if isEtcdKeyNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func isEtcdKeyNotFound(err error) bool {
	if cErr, ok := err.(etcd.Error); ok {
		return cErr.Code == etcd.ErrorCodeKeyNotFound
	}
	return false
}

type EtcdClusterInfo struct {
	Health  bool            `json:"health"`
	Members []MemberWrapper `json:"members"`
}

type MemberWrapper struct {
	etcd.Member
	Leader   bool   `json:"leader"`
	Health   bool   `json:"health"`
	EndPoint string `json:"endpoint"`
}

func (s *EtcdStore) ClusterInfo() (EtcdClusterInfo, error) {
	var eci = EtcdClusterInfo{}

	members, err := s.mapi.List(context.Background())
	if err != nil {
		return eci, err
	}

	leader, err := s.mapi.Leader(context.Background())
	if err != nil {
		return eci, err
	}

	h := func(u string) bool {
		u = strings.TrimSuffix(u, "/") + "/health"
		resp, err := http.Get(u)
		if err != nil {
			return false
		}
		var health struct {
			Health string `json:"health"`
		}
		json.NewDecoder(resp.Body).Decode(&health)
		return health.Health == "true"
	}

	var mws []MemberWrapper
	for _, member := range members {
		mw := MemberWrapper{
			Member:   member,
			Leader:   member.ID == leader.ID,
			Health:   h(member.ClientURLs[0]),
			EndPoint: member.ClientURLs[0],
		}
		if mw.Health {
			eci.Health = true
		}
		mws = append(mws, mw)
	}
	eci.Members = mws

	return eci, nil
}

// encode & decode is just short-hands for json Marshal/Unmarshal
func encode(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func decode(bs []byte, v interface{}) error {
	return json.Unmarshal(bs, v)
}
