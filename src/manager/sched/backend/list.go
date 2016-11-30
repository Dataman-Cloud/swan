package backend

import (
	"github.com/Dataman-Cloud/swan/src/types"
)

func (b *Backend) ListApplications() ([]*types.Application, error) {
	return b.store.ListApplications()
}

func (b *Backend) ListApplicationTasks(appId string) ([]*types.Task, error) {
	return b.store.ListTasks(appId)
}

// ListApplicationVersions is used to list all versions for application from db specified by application id.
func (b *Backend) ListApplicationVersions(appId string) ([]string, error) {
	return b.store.ListVersionId(appId)
}
