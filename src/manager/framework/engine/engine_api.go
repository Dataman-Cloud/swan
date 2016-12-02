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

	app, err := state.NewApp(version, engine.Allocator)
	if err != nil {
		return err
	}

	engine.appLock.Lock()
	defer engine.appLock.Unlock()

	engine.Apps[version.AppId] = app

	return nil
}

func (engine *Engine) UpdateApp(version *types.Version) error {
	app, appExists := engine.Apps[version.AppId]
	if !appExists {
		return errors.New("app not exists")
	}

	if err := app.Update(version); err != nil {
		return err
	}

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
