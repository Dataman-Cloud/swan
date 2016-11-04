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

	UpdateAppStatus(appId string, status string) error

	AddAppUpdatedInstance(appId string, increment int) error

	updateAppUpdatedInstance(appId string, instances int) error

	PutTasks(tasks ...*types.Task) error

	PutTask(task *types.Task) error

	UpdateTaskStatus(appId, taskId, status string) error

	GetTask(appId, taskId string) (*types.Task, error)

	GetTasks(appId string, taskIds ...string) ([]*types.Task, error)

	DeleteTask(appId, taskId string) error

	DeleteTasks(appId string, taskIds ...string) error

	PutHealthcheck(task *types.Task, port uint32, appId string) error

	GetHealthChecks(appId string) ([]*types.Check, error)

	DeleteHealthCheck(appId, healthCheckId string) error

	// RegisterApplicationVersion is used to register a application version in consul.
	RegisterApplicationVersion(string, *types.ApplicationVersion) error

	// ListApplicationVersions is used to list all version ids for application.
	ListApplicationVersions(string) ([]string, error)

	// FetchApplicationVersion is used to fetch specified version from consul by version id and application id.
	FetchApplicationVersion(string, string) (*types.ApplicationVersion, error)

	// UpdateApplication is used to update application info.
	UpdateApplication(string, string, string) error
}
