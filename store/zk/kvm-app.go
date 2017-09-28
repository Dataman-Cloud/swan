package zk

import (
	"fmt"
	"path"
	"strings"

	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
)

// Kvm App
//
//
func (zk *ZKStore) CreateKvmApp(app *types.KvmApp) error {
	p := path.Join(keyKvmApp, app.ID)

	data, _, err := zk.get(p)
	if err != nil {
		if !strings.Contains(err.Error(), "node does not exist") {
			log.Errorf("find app %s got error: %v", app.ID, err)
			return err
		}
	}

	if len(data) != 0 {
		return errAppAlreadyExists
	}

	bs, err := encode(app)
	if err != nil {
		return err
	}

	return zk.createAll(p, bs)
}

func (zk *ZKStore) UpdateKvmApp(app *types.KvmApp) error {
	bs, err := encode(app)
	if err != nil {
		return err
	}

	path := path.Join(keyKvmApp, app.ID)

	return zk.set(path, bs)
}

func (zk *ZKStore) GetKvmApp(id string) (*types.KvmApp, error) {
	p := path.Join(keyKvmApp, id)

	data, _, err := zk.get(p)
	if err != nil {
		log.Errorf("find app %s got error: %v", id, err)
		return nil, fmt.Errorf("find app %s got error: %v", id, err)
	}

	var app types.KvmApp
	if err := decode(data, &app); err != nil {
		return nil, err
	}

	tasks, err := zk.ListKvmTasks(id)
	if err != nil {
		log.Errorf("get app %s tasks got error: %v", id, err)
		return nil, err
	}

	app.TaskCount = len(tasks)
	app.TasksStatus = zk.kvmTasksStatus(tasks)

	return &app, nil
}

func (zk *ZKStore) ListKvmApps() ([]*types.KvmApp, error) {
	nodes, err := zk.list(keyKvmApp)
	if err != nil {
		log.Errorln("zk ListApps error:", err)
		return nil, err
	}

	apps := make([]*types.KvmApp, 0)
	for _, node := range nodes {
		app, err := zk.GetKvmApp(node)
		if err != nil {
			log.Errorf("get app error: %v", err)
			continue
		}
		apps = append(apps, app)
	}

	return apps, nil
}

func (zk *ZKStore) DeleteKvmApp(id string) error {
	p := path.Join(keyKvmApp, id)

	if err := zk.del(path.Join(keyKvmApp, id, "tasks")); err != nil {
		log.Errorf("delete app %s tasks key got error: %v", id, err)
		return err
	}

	return zk.del(p)
}

func (zk *ZKStore) kvmTasksStatus(tasks []*types.KvmTask) map[string]int {
	ret := make(map[string]int)
	for _, task := range tasks {
		ret[task.Status]++
	}
	return ret
}

// Kvm Task
//
//
func (zk *ZKStore) CreateKvmTask(aid string, task *types.KvmTask) error {
	bs, err := encode(task)
	if err != nil {
		return err
	}

	p := path.Join(keyKvmApp, aid, "tasks", task.ID)

	return zk.createAll(p, bs)
}

func (zk *ZKStore) UpdateKvmTask(aid string, task *types.KvmTask) error {
	bs, err := encode(task)
	if err != nil {
		return err
	}

	p := path.Join(keyKvmApp, aid, "tasks", task.ID)

	return zk.set(p, bs)
}

func (zk *ZKStore) ListKvmTasks(id string) ([]*types.KvmTask, error) {
	p := path.Join(keyKvmApp, id, "tasks")

	children, err := zk.list(p)
	if err != nil {
		log.Errorf("get kvm app %s children(tasks) error: %v", id, err)
		return nil, err
	}

	tasks := make([]*types.KvmTask, 0)
	for _, child := range children {
		p := path.Join(keyKvmApp, id, "tasks", child)
		data, _, err := zk.get(p)
		if err != nil {
			log.Errorf("get %s got error: %v", p, err)
			return nil, err
		}

		var t *types.KvmTask
		if err := decode(data, &t); err != nil {
			log.Errorf("decode kvm task %s got error: %v", id, err)
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (zk *ZKStore) DeleteKvmTask(id string) error {
	parts := strings.SplitN(id, ".", 3)
	p := path.Join(keyKvmApp, parts[2], "tasks", id)

	return zk.del(p)
}

func (zk *ZKStore) GetKvmTask(aid, tid string) (*types.KvmTask, error) {
	p := path.Join(keyKvmApp, aid, "tasks", tid)

	data, _, err := zk.get(p)
	if err != nil {
		if err == errNotExists {
			return nil, fmt.Errorf("kvm task %s not exist", tid)
		}

		return nil, err
	}

	var task types.KvmTask
	if err := decode(data, &task); err != nil {
		return nil, err
	}

	return &task, nil
}
