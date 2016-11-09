package backend

import (
	"github.com/Dataman-Cloud/swan/types"
)

// RegisterApplication register application in db.
func (b *Backend) SaveApplication(application *types.Application) error {
	return b.store.SaveApplication(application)
}

// RegisterApplicationVersion register application version in db.
func (b *Backend) SaveVersion(appId string, version *types.Version) error {
	return b.store.SaveVersion(version)
}
