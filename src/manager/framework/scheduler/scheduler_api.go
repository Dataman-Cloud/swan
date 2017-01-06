package scheduler

import (
	"errors"

	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/types"
)

func (scheduler *Scheduler) CreateApp(version *types.Version) (*state.App, error) {
	// TODO
	existedApp := scheduler.AppStorage.Get(version.AppID)
	if existedApp != nil {
		return nil, errors.New("app already exists")
	}

	app, err := state.NewApp(version, scheduler.UserEventChan)
	if err != nil {
		return nil, err
	}

	scheduler.AppStorage.Add(app.ID, app)

	return app, nil
}

func (scheduler *Scheduler) InspectApp(appId string) (*state.App, error) {
	app := scheduler.AppStorage.Get(appId)
	if app == nil {
		return nil, errors.New("app not exists")
	}
	return app, nil
}

func (scheduler *Scheduler) DeleteApp(appId string) error {
	app := scheduler.AppStorage.Get(appId)
	if app == nil {
		return errors.New("app not exists")
	}

	return app.Delete()
}

func (scheduler *Scheduler) ListApps(appFilterOptions AppFilterOptions) []*state.App {
	apps := make([]*state.App, 0)
	for _, v := range scheduler.AppStorage.Filter(appFilterOptions) {
		apps = append(apps, v)
	}

	return apps
}

func (scheduler *Scheduler) ScaleUp(appId string, newInstances int, newIps []string) error {
	app := scheduler.AppStorage.Get(appId)
	if app == nil {
		return errors.New("app not exists")
	}

	return app.ScaleUp(newInstances, newIps)
}

func (scheduler *Scheduler) ScaleDown(appId string, removeInstances int) error {
	app := scheduler.AppStorage.Get(appId)
	if app == nil {
		return errors.New("app not exists")
	}

	return app.ScaleDown(removeInstances)
}

func (scheduler *Scheduler) UpdateApp(appId string, version *types.Version) error {
	app := scheduler.AppStorage.Get(appId)
	if app == nil {
		return errors.New("app not exists")
	}

	return app.Update(version, scheduler.store)
}

func (scheduler *Scheduler) CancelUpdate(appId string) error {
	app := scheduler.AppStorage.Get(appId)
	if app == nil {
		return errors.New("app not exists")
	}

	return app.CancelUpdate()
}

func (scheduler *Scheduler) ProceedUpdate(appId string, instances int) error {
	app := scheduler.AppStorage.Get(appId)
	if app == nil {
		return errors.New("app not exists")
	}

	return app.ProceedingRollingUpdate(instances)
}
