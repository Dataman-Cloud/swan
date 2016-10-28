package backend

import (
	"github.com/Dataman-Cloud/swan/types"
)

func (b *Backend) FetchApplication(id string) (*types.Application, error) {
	return b.store.FetchApplication(id)
}

// FetchApplicationVersion is used to fetch specified version from consul by version id and application id.
func (b *Backend) FetchApplicationVersion(applicationId, versionId string) (*types.ApplicationVersion, error) {
	return b.store.FetchApplicationVersion(applicationId, versionId)
}
