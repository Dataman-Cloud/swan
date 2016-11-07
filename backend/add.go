package backend

import (
	"github.com/Dataman-Cloud/swan/types"
)

// RegisterApplication register application in consul.
func (b *Backend) SaveApplication(application *types.Application) error {
	return b.store.SaveApplication(application)
}

// RegisterApplicationVersion register application version in consul.
func (b *Backend) SaveVersion(appId string, version *types.Version) error {
	return b.store.SaveVersion(version)
}
