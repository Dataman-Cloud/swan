package mock

import "github.com/Dataman-Cloud/swan/types"

type Store struct{}

func (s *Store) RegisterFrameworkID(id string) error {
	return nil
}

func (s *Store) FrameworkIDHasRegistered(id string) (bool, error) {
	return true, nil
}

func (s *Store) RegisterApplication(app *types.Application) error {
	return nil
}

func (s *Store) ListApplications() ([]*types.Application, error) {
	return nil, nil
}

func (s *Store) FetchApplication(id string) (*types.Application, error) {
	return nil, nil
}

func (s *Store) DeleteApplication(id string) error {
	return nil
}

func (s *Store) RegisterTask(task *types.Task) error {
	return nil
}

func (s *Store) ListApplicationTasks(id string) ([]*types.Task, error) {
	return nil, nil
}

func (s *Store) DeleteApplicationTasks(id string) error {
	return nil
}

func (s *Store) FetchApplicationTask(appId, taskId string) (*types.Task, error) {
	return nil, nil
}

func (s *Store) DeleteApplicationTask(appId, taskId string) error {
	return nil
}

func (s *Store) RegisterApplicationVersion(appId string, version *types.Version) error {
	return nil
}

func (s *Store) ListApplicationVersions(appId string) ([]string, error) {
	return nil, nil
}

func (s *Store) FetchApplicationVersion(appId, verionId string) (*types.Version, error) {
	return nil, nil
}

func (s *Store) UpdateApplication(appId, key, value string) error {
	return nil
}

func (s *Store) RegisterCheck(task *types.Task, port uint32, appId string) error {
	return nil
}

func (s *Store) DeleteCheck(checkId string) error {
	return nil
}

func (s *Store) UpdateTask(appId, taskId, status string) error {
	return nil
}

func (s *Store) IncreaseApplicationInstances(appId string) error {
	return nil
}

func (s *Store) IncreaseApplicationUpdatedInstances(appId string) error {
	return nil
}

func (s *Store) ResetApplicationUpdatedInstances(appId string) error {
	return nil
}

func (s *Store) UpdateApplicationStatus(appId, status string) error {
	return nil
}

func (s *Store) IncreaseApplicationRunningInstances(appId string) error {
	return nil
}

func (s *Store) ReduceApplicationInstances(appId string) error {
	return nil
}
