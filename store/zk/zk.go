package zk

import (
	"encoding/json"
	"errors"
	"net/url"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
)

var (
	zs *ZKStore
)

var (
	errAppNotFound          = errors.New("app not found")
	errAppAlreadyExists     = errors.New("app already exists")
	errVersionAlreadyExists = errors.New("version already exists")
	errComposeNotFound      = errors.New("compose app not found")
	errNotExists            = zk.ErrNoNode
)

const (
	keyApp         = "/apps"        // single app
	keyCompose     = "/composes"    // compose instance legacy (group apps), deprecated
	keyComposeNG   = "/composes-ng" // compose instance (group apps)
	keyFrameworkID = "/frameworkId" // framework id
)

type ZKStore struct {
	url  *url.URL
	conn *zk.Conn
	acl  []zk.ACL
}

func DB() *ZKStore {
	return zs
}

func NewZKStore(url *url.URL) (*ZKStore, error) {
	zs = &ZKStore{
		url: url,
		acl: zk.WorldACL(zk.PermAll),
	}

	if err := zs.initConnection(); err != nil {
		return nil, err
	}

	// create base keys nodes
	for _, node := range []string{keyApp, keyCompose, keyComposeNG, keyFrameworkID} {
		if err := zs.ensure(node); err != nil {
			return nil, err
		}
	}

	return zs, nil
}

func (zs *ZKStore) IsErrNotFound(err error) bool {
	if err == nil {
		return false
	}
	switch err {
	case errComposeNotFound, errAppNotFound, errNotExists, zk.ErrNoNode:
		return true
	default:
		return strings.Contains(err.Error(), "node does not exist")
	}
}

func (zs *ZKStore) initConnection() error {
	hosts := strings.Split(zs.url.Host, ",")
	conn, connCh, err := zk.Connect(hosts, 5*time.Second)
	if err != nil {
		return err
	}

	// waiting for zookeeper to be connected.
	for event := range connCh {
		if event.State == zk.StateConnected {
			log.Info("connected to zookeeper succeed.")
			break
		}
	}

	zs.conn = conn
	return nil
}

func (zs *ZKStore) ensure(path string) error {
	if exists, _ := zs.exist(path); exists {
		return nil
	}
	return zs.createAll(path, nil)
}

// with the prefix `s.url.Path` and clean the path
func (zs *ZKStore) clean(p string) string {
	if !strings.HasPrefix(p, zs.url.Path) {
		p = zs.url.Path + "/" + p
	}
	return path.Clean(p)
}

func (zs *ZKStore) get(path string) (data []byte, stat *zk.Stat, err error) {
	data, stat, err = zs.conn.Get(zs.clean(path))
	return
}

func (zs *ZKStore) del(path string) error {
	exist, err := zs.exist(path)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	return zs.conn.Delete(zs.clean(path), -1)
}

func (zs *ZKStore) list(path string) (children []string, err error) {
	children, _, err = zs.conn.Children(zs.clean(path))
	return
}

func (zs *ZKStore) exist(path string) (exist bool, err error) {
	exist, _, err = zs.conn.Exists(zs.clean(path))
	return
}

func (zs *ZKStore) createAll(path string, data []byte) error {
	path = zs.clean(path)

	var (
		fields = strings.Split(path, "/")
		node   = "/"
	)

	// all of dir node
	for i, v := range fields[1:] {
		node += v
		if i >= len(fields[1:])-1 {
			break // the end node
		}
		err := zs.create(node, nil)
		if err != nil {
			log.Errorf("create node: %s error: %v", node, err)
			return err
		}
		node += "/"
	}

	// the end data node
	return zs.create(node, data)
}

func (zs *ZKStore) set(path string, data []byte) error {
	path = zs.clean(path)
	_, err := zs.conn.Set(path, data, -1)

	return err
}

func (zs *ZKStore) create(path string, data []byte) error {
	path = zs.clean(path)

	exist, err := zs.exist(path)
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	_, err = zs.conn.Create(path, data, 0, zs.acl)

	return err
}

// encode & decode is just short-hands for json Marshal/Unmarshal
func encode(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func decode(bs []byte, v interface{}) error {
	return json.Unmarshal(bs, v)
}
