package backend

import (
	"github.com/Dataman-Cloud/swan/types"
)

func (b *Backend) ListApplications() ([]*types.Application, error) {
	return b.store.GetApps()
}

func (b *Backend) ListApplicationTasks(id string) ([]*types.Task, error) {
	return b.store.GetTasks(id)
}

// ListApplicationVersions is used to list all versions for application from consul specified by application id.
func (b *Backend) ListApplicationVersions(applicationId string) ([]*types.ApplicationVersion, error) {
	return b.store.GetVersions(applicationId)
}
