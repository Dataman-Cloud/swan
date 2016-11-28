package backend

import (
	"github.com/Dataman-Cloud/swan/src/types"
)

func (b *Backend) FetchApplication(id string) (*types.Application, error) {
	return b.store.FetchApplication(id)
}

// FetchApplicationVersion is used to fetch specified version from db by version id and application id.
func (b *Backend) FetchApplicationVersion(applicationId, versionId string) (*types.Version, error) {
	return b.store.FetchVersion(versionId)
}
