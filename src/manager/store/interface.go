package store

import (
	"github.com/Dataman-Cloud/swan/src/types"
)

type Store interface {

	// framework

	// save framework id to db
	SaveFrameworkID(string) error

	// fetch framework id from db
	FetchFrameworkID() (string, error)

	// check if framework id is in db or not
	HasFrameworkID() (bool, error)

	// application

	// save application in db
	SaveApplication(*types.Application) error

	// fetch application from db
	FetchApplication(string) (*types.Application, error)

	// list all applications
	ListApplications() ([]*types.Application, error)

	// delete application fro db
	DeleteApplication(string) error

	// increase application updated instances count +1
	IncreaseApplicationUpdatedInstances(string) error

	// reset application updated instances to zero
	ResetApplicationUpdatedInstances(string) error

	// increase application instances
	IncreaseApplicationInstances(string) error

	// reduce application instances
	ReduceApplicationInstances(string) error

	// update application status
	UpdateApplicationStatus(string, string) error

	// increase application running instances count
	IncreaseApplicationRunningInstances(string) error

	// reduce application running instances count
	ReduceApplicationRunningInstances(string) error

	// task

	// save task in db
	SaveTask(*types.Task) error

	// list all tasks belong to a application
	ListTasks(appId string) ([]*types.Task, error)

	// fetch task from db
	FetchTask(appId, taskId string) (*types.Task, error)

	// delete task from db
	DeleteTask(appId, taskId string) error

	// update task status
	UpdateTaskStatus(appId, taskId, status string) error

	// version

	// save version to db
	SaveVersion(*types.Version) error

	// list all versions
	ListVersionId(appId string) ([]string, error)

	// fetch version from db by version id
	FetchVersion(appId, versionId string) (*types.Version, error)

	// delete version from db
	DeleteVersion(appId, versionId string) error
}
