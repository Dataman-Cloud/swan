package store

import "github.com/Dataman-Cloud/swan/src/types"

func (store *ManagerStore) SaveApplication(application *types.Application) error {
	return nil
}

func (store *ManagerStore) FetchApplication(appId string) (*types.Application, error) {
	return nil, nil
}

func (store *ManagerStore) ListApplications() ([]*types.Application, error) {
	return nil, nil
}

func (store *ManagerStore) DeleteApplication(appId string) error {
	return nil
}

func (store *ManagerStore) IncreaseApplicationUpdatedInstances(appId string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.UpdatedInstances += 1

	return store.SaveApplication(app)
}

func (store *ManagerStore) IncreaseApplicationInstances(appId string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.Instances += 1

	return store.SaveApplication(app)
}

func (store *ManagerStore) ResetApplicationUpdatedInstances(appId string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.UpdatedInstances = 0

	return store.SaveApplication(app)
}

func (store *ManagerStore) UpdateApplicationStatus(appId, status string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.Status = status

	return store.SaveApplication(app)
}

func (store *ManagerStore) IncreaseApplicationRunningInstances(appId string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.RunningInstances += 1

	return store.SaveApplication(app)
}

func (store *ManagerStore) ReduceApplicationRunningInstances(appId string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.RunningInstances -= 1

	return store.SaveApplication(app)
}

func (store *ManagerStore) ReduceApplicationInstances(appId string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.Instances -= 1

	return store.SaveApplication(app)
}
