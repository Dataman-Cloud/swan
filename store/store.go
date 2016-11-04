package store

import "github.com/Dataman-Cloud/swan/types"

type Store interface {
	PutFrameworkID(string) error

	GetFrameworkID() (string, error)

	PutApp(*types.Application) error

	PutApps(...*types.Application) error

	GetApp(string) (*types.Application, error)

	GetApps(...string) ([]*types.Application, error)

	DeleteApp(string) error

	DeleteApps(...string) error

	AddAppInstance(appId string, increment int) error

	AddAppRunningInstance(appId string, increment int) error

	UpdateAppStatus(appId string, status string) error

	AddAppUpdatedInstance(appId string, increment int) error

	UpdateAppUpdatedInstance(appId string, instances int) error

	PutTasks(appId string, tasks ...*types.Task) error

	PutTask(appId string, task *types.Task) error

	UpdateTaskStatus(appId, taskId, status string) error

	GetTask(appId, taskId string) (*types.Task, error)

	GetTasks(appId string, taskIds ...string) ([]*types.Task, error)

	DeleteTask(appId, taskId string) error

	DeleteTasks(appId string, taskIds ...string) error

	PutHealthcheck(task *types.Task, port uint32, appId string) error

	GetHealthChecks(appId string) ([]*types.Check, error)

	DeleteHealthCheck(appId, healthCheckId string) error

	PutVersions(appId string, versions ...*types.ApplicationVersion) error

	PutVersion(appId string, version *types.ApplicationVersion) error

	GetVersions(appId string, versionIds ...string) ([]*types.ApplicationVersion, error)

	GetAndSortVersions(appId string, versionIds ...string) ([]*types.ApplicationVersion, error)

	GetVersion(appId, versionId string) (*types.ApplicationVersion, error)

	DeleteVersions(appId string, versionIds ...string) error

	DeleteVersion(appId, versionId string) error
}
