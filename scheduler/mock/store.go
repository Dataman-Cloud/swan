package mock

import "github.com/Dataman-Cloud/swan/types"

type Store struct{}

func (s *Store) SaveFrameworkID(id string) error {
	return nil
}

func (s *Store) HasFrameworkID(id string) (bool, error) {
	return true, nil
}

func (s *Store) FetchFrameworkID(key string) (string, error) {
	return "xxxxx", nil
}

func (s *Store) SaveApplication(app *types.Application) error {
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

func (s *Store) SaveTask(task *types.Task) error {
	return nil
}

func (s *Store) ListTasks(id string) ([]*types.Task, error) {
	return nil, nil
}

func (s *Store) DeleteTasks(id string) error {
	return nil
}

func (s *Store) FetchTask(taskId string) (*types.Task, error) {
	return nil, nil
}

func (s *Store) DeleteTask(taskId string) error {
	return nil
}

func (s *Store) SaveVersion(version *types.Version) error {
	return nil
}

func (s *Store) ListVersions(appId string) ([]string, error) {
	return nil, nil
}

func (s *Store) FetchVersion(verionId string) (*types.Version, error) {
	return nil, nil
}

func (s *Store) UpdateApplication(appId, key, value string) error {
	return nil
}

func (s *Store) DeleteVersion(versionId string) error {
	return nil
}

func (s *Store) SaveCheck(task *types.Task, port uint32, appId string) error {
	return nil
}

func (s *Store) DeleteCheck(checkId string) error {
	return nil
}

func (s *Store) UpdateTaskStatus(taskId, status string) error {
	return nil
}

func (s *Store) ListChecks() ([]*types.Check, error) {
	return nil, nil
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

func (s *Store) ReduceApplicationRunningInstances(appId string) error {
	return nil
}

func (s *Store) ReduceApplicationInstances(appId string) error {
	return nil
}
