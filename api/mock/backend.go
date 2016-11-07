package mock

import (
	"github.com/Dataman-Cloud/swan/types"
)

type Backend struct{}

func (b *Backend) ClusterId() string {
	return "testId"
}

func (b *Backend) SaveApplication(app *types.Application) error {
	return nil
}

func (b *Backend) SaveVersion(app string, version *types.Version) error {
	return nil
}

func (b *Backend) LaunchApplication(version *types.Version) error {
	return nil
}

func (b *Backend) DeleteApplication(appId string) error {
	return nil
}

func (b *Backend) DeleteApplicationTasks(appId string) error {
	return nil
}

func (b *Backend) ListApplications() ([]*types.Application, error) {
	return nil, nil
}

func (b *Backend) FetchApplication(appId string) (*types.Application, error) {
	return nil, nil
}

func (b *Backend) ListApplicationTasks(appId string) ([]*types.Task, error) {
	return nil, nil
}

func (b *Backend) DeleteApplicationTask(appId string, taskId string) error {
	return nil
}

func (b *Backend) ListApplicationVersions(appId string) ([]string, error) {
	return nil, nil
}

func (b *Backend) FetchApplicationVersion(appId string, versionId string) (*types.Version, error) {
	return nil, nil
}

func (b *Backend) UpdateApplication(string, int, *types.Version) error {
	return nil
}

func (b *Backend) ScaleApplication(appId string, instances int) error {
	return nil
}

func (b *Backend) RollbackApplication(appId string) error {
	return nil
}
