package store

import (
	"errors"
	"net/url"

	"github.com/Dataman-Cloud/swan/store/etcd"
	"github.com/Dataman-Cloud/swan/store/zk"
	"github.com/Dataman-Cloud/swan/types"
)

type Store interface {
	CreateApp(app *types.Application) error
	UpdateApp(app *types.Application) error
	GetApp(appId string) (*types.Application, error)
	ListApps() ([]*types.Application, error)
	DeleteApp(appId string) error
	GetAppOpStatus(string) (string, error)

	CreateTask(string, *types.Task) error
	GetTask(string, string) (*types.Task, error)
	UpdateTask(string, *types.Task) error
	DeleteTask(string) error
	ListTasks(string) ([]*types.Task, error)

	CreateVersion(string, *types.Version) error
	GetVersion(string, string) (*types.Version, error)
	ListVersions(string) ([]*types.Version, error)
	DeleteVersion(string, string) error

	UpdateFrameworkId(frameworkId string) error
	GetFrameworkId() (string, int64)

	// compose legacy, Deprecated.
	CreateCompose(ins *types.Compose) error
	DeleteCompose(id string) error
	UpdateCompose(ins *types.Compose) error // status, errmsg, updateAt
	GetCompose(id string) (*types.Compose, error)
	ListComposes() ([]*types.Compose, error)

	// compose ng
	CreateComposeNG(cmpApp *types.ComposeApp) error
	DeleteComposeNG(idOrName string) error
	UpdateComposeNG(cmpApp *types.ComposeApp) error // status, errmsg, updateAt
	GetComposeNG(idOrName string) (*types.ComposeApp, error)
	ListComposesNG() ([]*types.ComposeApp, error)

	IsErrNotFound(err error) bool

	// mesos agent labels
	CreateMesosAgent(*types.MesosAgent) error
	GetMesosAgent(string) (*types.MesosAgent, error)
	UpdateMesosAgent(*types.MesosAgent) error

	// mesos virtual cluster
	ListVClusters() ([]*types.VCluster, error)
	CreateVCluster(*types.VCluster) error
	GetVCluster(string) (*types.VCluster, error)
	VClusterExists(string) bool
}

func Setup(typ string, zkURL *url.URL, etcdAddrs []string) (Store, error) {
	switch typ {
	case "zk":
		return zk.NewZKStore(zkURL)
	case "etcd":
		return etcd.NewEtcdStore(etcdAddrs)
	}

	return nil, errors.New("unsuported db store type: " + typ)
}
