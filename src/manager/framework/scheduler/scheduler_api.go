package scheduler

import (
	"errors"

	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/types"
)

func (scheduler *Scheduler) CreateApp(version *types.Version) (*state.App, error) {
	_, appExists := scheduler.Apps[version.AppId]
	if appExists {
		return nil, errors.New("app already exists")
	}

	app, err := state.NewApp(version, scheduler.Allocator, scheduler.MesosConnector)
	if err != nil {
		return nil, err
	}

	scheduler.appLock.Lock()
	defer scheduler.appLock.Unlock()

	scheduler.Apps[version.AppId] = app

	return app, nil
}

func (scheduler *Scheduler) InspectApp(appId string) (*state.App, error) {
	app, appExists := scheduler.Apps[appId]
	if !appExists {
		return nil, errors.New("app not exists")
	}
	return app, nil
}

func (scheduler *Scheduler) DeleteApp(appId string) error {
	app, appExists := scheduler.Apps[appId]
	if !appExists {
		return errors.New("app not exists")
	}
	return app.Delete()
}

func (scheduler *Scheduler) ListApps() []*state.App {
	apps := make([]*state.App, 0)
	for _, v := range scheduler.Apps {
		apps = append(apps, v)
	}

	return apps
}

func (scheduler *Scheduler) ScaleUp(appId string, to int) error {
	app, appExists := scheduler.Apps[appId]
	if !appExists {
		return errors.New("app not exists")
	}
	return app.ScaleUp(to)
}

func (scheduler *Scheduler) ScaleDown(appId string, to int) error {
	app, appExists := scheduler.Apps[appId]
	if !appExists {
		return errors.New("app not exists")
	}
	return app.ScaleDown(to)
}

func (scheduler *Scheduler) UpdateApp(appId string, version *types.Version) error {
	app, appExists := scheduler.Apps[appId]
	if !appExists {
		return errors.New("app doesn't exists, update failed")
	}

	return app.Update(version)
}

func (scheduler *Scheduler) CancelUpdate(appId string) error {
	app, appExists := scheduler.Apps[appId]
	if !appExists {
		return errors.New("app doesn't exists, update failed")
	}

	return app.CancelUpdate()
}

func (scheduler *Scheduler) ProceedUpdate(appId string, instances int) error {
	app, appExists := scheduler.Apps[appId]
	if !appExists {
		return errors.New("app doesn't exists, update failed")
	}

	return app.ProceedingRollingUpdate(instances)
}
