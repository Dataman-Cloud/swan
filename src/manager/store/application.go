package store

func (zk *ZkStore) CreateApp(app *Application) error {
	op := &StoreOp{
		Op:      OP_ADD,
		Entity:  ENTITY_APP,
		Param1:  app.ID,
		Payload: app,
	}

	return zk.Apply(op)
}

func (zk *ZkStore) UpdateApp(app *Application) error {
	op := &StoreOp{
		Op:      OP_UPDATE,
		Entity:  ENTITY_APP,
		Param1:  app.ID,
		Payload: app.Bytes(),
	}

	return zk.Apply(op)
}

func (zk *ZkStore) GetApp(appId string) *Application {
	appStore, found := zk.Apps[appId]
	if !found {
		return nil
	}

	return appStore.App
}

func (zk *ZkStore) ListApps() []*Application {
	apps := make([]*Application, 0)
	for _, app := range zk.Apps {
		apps = append(apps, app.App)
	}

	return apps
}

func (zk *ZkStore) DeleteApp(appId string) error {
	op := &StoreOp{
		Op:     OP_REMOVE,
		Entity: ENTITY_APP,
		Param1: appId,
	}

	return zk.Apply(op)
}
