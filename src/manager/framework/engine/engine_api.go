package engine

import (
	"errors"

	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/types"
)

func (engine *Engine) CreateApp(version *types.Version) error {
	_, appExists := engine.Apps[version.AppId]
	if appExists {
		return errors.New("app already exists")
	}

	app, err := state.NewApp(version, engine.Allocator, engine.Scheduler)
	if err != nil {
		return err
	}

	engine.appLock.Lock()
	defer engine.appLock.Unlock()

	engine.Apps[version.AppId] = app

	return nil
}

func (engine *Engine) InspectApp(appId string) (*state.App, error) {
	app, appExists := engine.Apps[appId]
	if !appExists {
		return nil, errors.New("app not exists")
	}
	return app, nil
}

func (engine *Engine) DeleteApp(appId string) error {
	app, appExists := engine.Apps[appId]
	if !appExists {
		return errors.New("app not exists")
	}
	return app.Delete()
}

func (engine *Engine) ListApps() []*state.App {
	apps := make([]*state.App, 0)
	for _, v := range engine.Apps {
		apps = append(apps, v)
	}

	return apps
}

func (engine *Engine) ScaleUp(appId string, to int) error {
	app, appExists := engine.Apps[appId]
	if !appExists {
		return errors.New("app not exists")
	}
	return app.ScaleUp(to)
}

func (engine *Engine) ScaleDown(appId string, to int) error {
	app, appExists := engine.Apps[appId]
	if !appExists {
		return errors.New("app not exists")
	}
	return app.ScaleDown(to)
}

func (engine *Engine) UpdateApp(appId string, version *types.Version) error {
	app, appExists := engine.Apps[appId]
	if !appExists {
		return errors.New("app doesn't exists, update failed")
	}

	return app.Update(version)
}

func (engine *Engine) CancelUpdate(appId string) error {
	app, appExists := engine.Apps[appId]
	if !appExists {
		return errors.New("app doesn't exists, update failed")
	}

	return app.CancelUpdate()
}

func (engine *Engine) ProceedUpdate(appId string, instances int) error {
	app, appExists := engine.Apps[appId]
	if !appExists {
		return errors.New("app doesn't exists, update failed")
	}

	return app.ProceedingRollingUpdate(instances)
}
