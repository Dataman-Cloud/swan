package backend

import (
	"github.com/Dataman-Cloud/swan/types"
)

// RegisterApplication register application in consul.
func (b *Backend) RegisterApplication(application *types.Application) error {
	return b.store.PutApp(application)
}

// RegisterApplicationVersion register application version in consul.
<<<<<<< d4131a5d2b159499b83ae9848fe54f1aa6b14d75
func (b *Backend) RegisterApplicationVersion(appId string, version *types.Version) error {
	return b.store.RegisterApplicationVersion(appId, version)
=======
func (b *Backend) RegisterApplicationVersion(appId string, version *types.ApplicationVersion) error {
	return b.store.PutVersion(appId, version)
>>>>>>> update backend
}
