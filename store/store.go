package store

import "github.com/Dataman-Cloud/swan/types"

type Store interface {
	PutFrameworkID(string) error

	GetFrameworkID(string) (string, error)

	PutApp(*types.Application) error

	PutApps(...*types.Application) error

	GetApp(string) (*types.Application, error)

	GetApps(...string) ([]*types.Application, error)

	DeleteApp(string) error

	DeleteApps(...string) error

	AddAppInstance(appId string, increment int) error

	AddAppRunningInstance(appId string, increment int) error

	PutAppStatus(appId string, status string) error

	AddAppUpdatedInstance(appId string, increment int) error

	SetAppUpdatedInstance(appId string, instances int) error

	// RegisterTask is used to register task in consul under task.AppId namespace.
	RegisterTask(*types.Task) error

	// ListApplicationTasks is used to get all tasks belong to a application from consul.
	ListApplicationTasks(string) ([]*types.Task, error)

	// DeleteApplicationTasks is used to delete all tasks belong to a application from consul.
	DeleteApplicationTasks(string) error

	// FetchApplicationTask is used to fetch a task belong to a application from consul.
	FetchApplicationTask(string, string) (*types.Task, error)

	// DeleteApplicationTask is used to delete specified task belong to a application from consul.
	DeleteApplicationTask(string, string) error

	// RegisterApplicationVersion is used to register a application version in consul.
	RegisterApplicationVersion(string, *types.ApplicationVersion) error

	// ListApplicationVersions is used to list all version ids for application.
	ListApplicationVersions(string) ([]string, error)

	// FetchApplicationVersion is used to fetch specified version from consul by version id and application id.
	FetchApplicationVersion(string, string) (*types.ApplicationVersion, error)

	// UpdateApplication is used to update application info.
	UpdateApplication(string, string, string) error

	// RegisterCheck register check in consul.
	RegisterCheck(*types.Task, uint32, string) error

	// DeleteCheck delete task health check from consul.
	DeleteCheck(string) error

	// UpdateTask updated task status by task id.
	UpdateTask(string, string, string) error
}
