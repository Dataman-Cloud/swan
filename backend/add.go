package backend

import (
	"github.com/Dataman-Cloud/swan/types"
)

// RegisterApplication register application in consul.
func (b *Backend) RegisterApplication(application *types.Application) error {
	return b.store.PutApp(application)
}

// RegisterApplicationVersion register application version in consul.
func (b *Backend) RegisterApplicationVersion(appId string, version *types.ApplicationVersion) error {
	return b.store.PutVersion(appId, version)
}
