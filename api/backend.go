package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/types"
)

type Backend interface {
	ClusterId() string
	// RegisterApplication register application in consul.
	RegisterApplication(*types.Application) error

	// RegisterApplicationVersion register application version in consul.
	RegisterApplicationVersion(string, *types.ApplicationVersion) error

	// LaunchApplication launch applications
	LaunchApplication(*types.ApplicationVersion) error

	// DeleteApplication will delete all data associated with application.
	DeleteApplication(string) error

	// DeleteApplicationTasks delete all tasks belong to appcaiton but keep that application exists.
	DeleteApplicationTasks(string) error

	EventStream(http.ResponseWriter) error

	ListApplications() ([]*types.Application, error)

	FetchApplication(string) (*types.Application, error)

	ListApplicationTasks(string) ([]*types.Task, error)

	DeleteApplicationTask(string, string) error

	ListApplicationVersions(string) ([]string, error)

	FetchApplicationVersion(string, string) (*types.ApplicationVersion, error)

	UpdateApplication(string, int, *types.ApplicationVersion) error

	ScaleApplication(string, int) error

	RollbackApplication(string) error
}
