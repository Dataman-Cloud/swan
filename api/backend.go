package api

import (
	"github.com/Dataman-Cloud/swan/types"
)

type Backend interface {
	ClusterId() string
	// RegisterApplication register application in consul.
	RegisterApplication(*types.Application) error

	// RegisterApplicationVersion register application version in consul.
	RegisterApplicationVersion(string, *types.Version) error

	// LaunchApplication launch applications
	LaunchApplication(*types.Version) error

	// DeleteApplication will delete all data associated with application.
	DeleteApplication(string) error

	// DeleteApplicationTasks delete all tasks belong to appcaiton but keep that application exists.
	DeleteApplicationTasks(string) error

	ListApplications() ([]*types.Application, error)

	FetchApplication(string) (*types.Application, error)

	ListApplicationTasks(string) ([]*types.Task, error)

	DeleteApplicationTask(string, string) error

	ListApplicationVersions(string) ([]string, error)

	FetchApplicationVersion(string, string) (*types.Version, error)

	UpdateApplication(string, int64, *types.Version) error

	ScaleApplication(string, int64) error

	RollbackApplication(string) error
}
