package store

import "github.com/Sirupsen/logrus"

func (zk *ZKStore) CreateApp(app *Application) error {
	if zk.GetApp(app.ID) != nil {
		return errAppAlreadyExists
	}

	holder := AppHolder{
		App:      app,
		Versions: make(map[string]*Version),
		Slots:    make(map[string]*Slot),
	}

	bs, err := encode(holder)
	if err != nil {
		return err
	}

	path := keyApp + "/" + app.ID
	return zk.createAll(path, bs)
}

// All of AppHolder Write Ops Requires Transaction Lock
func (zk *ZKStore) UpdateApp(app *Application) error {
	zk.Lock()
	defer zk.Unlock()

	holder := zk.GetAppHolder(app.ID)
	if holder == nil {
		return errAppNotFound
	}

	*holder.App = *app

	bs, err := encode(holder)
	if err != nil {
		return err
	}

	path := keyApp + "/" + app.ID
	return zk.create(path, bs)
}

func (zk *ZKStore) GetAppHolder(id string) *AppHolder {
	bs, err := zk.get(keyApp + "/" + id)
	if err != nil {
		logrus.Errorln("zk GetAppHolder error:", err)
		return nil
	}

	holder := new(AppHolder)
	if err := decode(bs, &holder); err != nil {
		logrus.Errorln("zk GetAppHolder.decode error:", err)
		return nil
	}

	return holder
}

func (zk *ZKStore) GetApp(id string) *Application {
	holder := zk.GetAppHolder(id)
	if holder != nil && holder.App != nil {
		return holder.App
	}
	return nil
}

func (zk *ZKStore) ListApps() []*Application {
	ret := make([]*Application, 0, 0)

	nodes, err := zk.list(keyApp)
	if err != nil {
		logrus.Errorln("zk ListApps error:", err)
		return ret
	}

	for _, node := range nodes {
		if app := zk.GetApp(node); app != nil {
			ret = append(ret, app)
		}
	}

	return ret
}

// All of AppHolder Write Ops Requires Transaction Lock
func (zk *ZKStore) DeleteApp(id string) error {
	zk.Lock()
	defer zk.Unlock()

	if zk.GetApp(id) == nil {
		return errAppNotFound
	}

	return zk.del(keyApp + "/" + id)
}
