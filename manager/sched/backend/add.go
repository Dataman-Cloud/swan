package backend

import (
	"github.com/Dataman-Cloud/swan/types"
	"reflect"
	"sort"
)

// RegisterApplication register application in db.
func (b *Backend) SaveApplication(application *types.Application) error {
	return b.store.SaveApplication(application)
}

// RegisterApplicationVersion register application version in db.
func (b *Backend) SaveVersion(appId string, version *types.Version) error {
	versions, err := b.store.ListVersions(appId)
	if err != nil {
		return err
	}

	if len(versions) != 0 {
		sort.Strings(versions)

		newestVersion, err := b.store.FetchVersion(versions[len(versions)-1])
		if err != nil {
			return err
		}

		if reflect.DeepEqual(version, newestVersion) {
			return nil
		}

	}

	return b.store.SaveVersion(version)
}
