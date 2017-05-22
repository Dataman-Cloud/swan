package store

import (
	"encoding/json"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
)

var (
	zs *ZKStore
)

type ZKStore struct {
	url  *url.URL
	conn *zk.Conn
	acl  []zk.ACL

	sync.Mutex // only protect AppHolder Nested Objects (Tasks,Slots ...) Write Ops
}

func DB() *ZKStore {
	return zs
}

func InitZKStore(url *url.URL) error {
	zs = &ZKStore{
		url: url,
		acl: zk.WorldACL(zk.PermAll),
	}

	if err := zs.initConnection(); err != nil {
		return err
	}

	// create base keys nodes
	for _, node := range []string{keyApp, keyInstance, keyAllocator} {
		if err := zs.createAll(node, nil); err != nil {
			return err
		}
	}

	return nil
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

// with the prefix `s.url.Path` and clean the path
func (zs *ZKStore) clean(p string) string {
	if !strings.HasPrefix(p, zs.url.Path) {
		p = zs.url.Path + "/" + p
	}
	return path.Clean(p)
}

func (zs *ZKStore) get(path string) (data []byte, err error) {
	data, _, err = zs.conn.Get(zs.clean(path))
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

func (zs *ZKStore) create(path string, data []byte) error {
	path = zs.clean(path)

	exist, err := zs.exist(path)
	if err != nil {
		return err
	}

	if exist {
		_, err = zs.conn.Set(path, data, -1)
	} else {
		_, err = zs.conn.Create(path, data, 0, zs.acl)
	}

	return err
}

// encode & decode is just short-hands for json Marshal/Unmarshal
func encode(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}
func decode(bs []byte, v interface{}) error {
	return json.Unmarshal(bs, v)
}
