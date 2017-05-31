package store

import (
	"github.com/Dataman-Cloud/swan/types"
)

type Store interface {
	CreateApp(app *types.Application) error
	UpdateApp(app *types.Application) error
	GetApp(appId string) (*types.Application, error)
	ListApps() ([]*types.Application, error)
	DeleteApp(appId string) error

	CreateTask(string, *types.Task) error
	GetTask(string, string) (*types.Task, error)
	UpdateTask(string, *types.Task) error
	DeleteTask(string) error
	ListTasks(string) ([]*types.Task, error)

	CreateVersion(string, *types.Version) error
	GetVersion(string, string) (*types.Version, error)
	ListVersions(string) ([]*types.Version, error)

	UpdateFrameworkId(frameworkId string) error
	GetFrameworkId() string

	CreateCompose(ins *types.Compose) error
	DeleteCompose(idOrName string) error
	UpdateCompose(ins *types.Compose) error // status, errmsg, updateAt
	GetCompose(idOrName string) (*types.Compose, error)
	ListComposes() ([]*types.Compose, error)

	GetLeader() string
}
