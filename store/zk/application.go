package zk

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
)

func (zk *ZKStore) CreateApp(app *types.Application) error {
	p := path.Join(keyApp, app.ID)

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

// All of AppHolder Write Ops Requires Transaction Lock
func (zk *ZKStore) UpdateApp(app *types.Application) error {
	bs, err := encode(app)
	if err != nil {
		return err
	}

	path := path.Join(keyApp, app.ID)

	return zk.set(path, bs)
}

func (zk *ZKStore) GetApp(id string) (*types.Application, error) {
	p := path.Join(keyApp, id)

	data, _, err := zk.get(p)
	if err != nil {
		log.Errorf("find app %s got error: %v", id, err)
		return nil, fmt.Errorf("find app %s got error: %v", id, err)
	}

	var app types.Application
	if err := decode(data, &app); err != nil {
		return nil, err
	}

	tasks, err := zk.tasks(p, id)
	if err != nil {
		log.Errorf("get app %s tasks got error: %v", id, err)
		return nil, err
	}

	app.TaskCount = len(tasks)
	app.Status = zk.status(tasks)
	app.TasksStatus = zk.tasksStatus(tasks)
	app.Version = zk.version(tasks)
	app.Health = zk.health(tasks)
	app.Progress, app.ProgressDetails = zk.progress(tasks)

	versions, err := zk.versions(p, id)
	if err != nil {
		log.Errorf("get app %s versions got error: %v", id, err)
		return nil, err
	}

	if len(app.Version) == 0 {
		if len(versions) > 0 {
			types.VersionList(versions).Reverse()
			app.Version = append(app.Version, versions[0].ID)
		}
	}

	app.VersionCount = len(versions)

	return &app, nil
}

func (zk *ZKStore) GetAppOpStatus(appId string) (string, error) {
	p := path.Join(keyApp, appId)

	data, _, err := zk.get(p)
	if err != nil {
		log.Errorf("find app %s got error: %v", appId, err)
		return "", fmt.Errorf("find app %s got error: %v", appId, err)
	}

	var app types.Application
	if err := decode(data, &app); err != nil {
		return "", err
	}

	return app.OpStatus, nil
}

func (zk *ZKStore) ListApps() ([]*types.Application, error) {
	nodes, err := zk.list(keyApp)
	if err != nil {
		log.Errorln("zk ListApps error:", err)
		return nil, err
	}

	apps := make([]*types.Application, 0)
	for _, node := range nodes {
		app, err := zk.GetApp(node)
		if err != nil {
			log.Errorf("get app error: %v", err)
			continue
		}
		apps = append(apps, app)
	}

	return apps, nil
}

func (zk *ZKStore) DeleteApp(id string) error {
	p := path.Join(keyApp, id)

	if err := zk.del(path.Join(keyApp, id, "tasks")); err != nil {
		log.Errorf("delete app %s tasks key got error: %v", id, err)
		return err
	}

	if err := zk.del(path.Join(keyApp, id, "versions")); err != nil {
		log.Errorf("delete app %s versions key got error: %v", id, err)
		return err
	}

	//if err := zk.delTasks(p, id); err != nil {
	//	log.Errorf("delete app %s tasks got error: %v", id, err)
	//	return err
	//}

	//if err := zk.delVersions(p, id); err != nil {
	//	log.Errorf("delete app %s versions got error: %v", id, err)
	//	return err
	//}

	return zk.del(p)
}

func (zk *ZKStore) tasks(p, id string) (types.TaskList, error) {
	children, err := zk.list(path.Join(p, "tasks"))
	if err != nil {
		log.Errorf("get app %s children(tasks) error: %v", id, err)
		return nil, err
	}

	tasks := make([]*types.Task, 0)
	for _, child := range children {
		p := path.Join(keyApp, id, "tasks", child)
		data, _, err := zk.get(p)
		if err != nil {
			continue
		}

		var task *types.Task
		if err := decode(data, &task); err != nil {
			log.Errorf("decode task %s got error: %v", id, err)
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (zk *ZKStore) versions(p, id string) (types.VersionList, error) {
	children, err := zk.list(path.Join(p, "versions"))
	if err != nil {
		log.Errorf("get app %s children(versions) error: %v", id, err)
		return nil, err
	}

	versions := make([]*types.Version, 0)
	for _, child := range children {
		p := path.Join(keyApp, id, "versions", child)
		data, _, err := zk.get(p)
		if err != nil {
			log.Errorf("get %s got error: %v", p, err)
			return nil, err
		}

		var ver *types.Version
		if err := decode(data, &ver); err != nil {
			log.Errorf("decode version %s got error: %v", id, err)
			return nil, err
		}

		versions = append(versions, ver)
	}

	return versions, err
}

func (zk *ZKStore) status(tasks types.TaskList) string {
	for _, task := range tasks {
		if task.Status == "TASK_RUNNING" {
			return "available"
		}
	}

	return "unavailable"
}

func (zk *ZKStore) tasksStatus(tasks types.TaskList) map[string]int {
	ret := make(map[string]int)
	for _, task := range tasks {
		ret[task.Status]++
	}
	return ret
}

func (zk *ZKStore) progress(tasks types.TaskList) (int, map[string]bool) {
	versions := zk.version(tasks)
	if len(versions) < 2 {
		return -1, nil
	}

	var (
		n int
		m = map[string]bool{}
	)
	for _, task := range tasks {
		if task.Version == versions[1] {
			n++
			m[task.ID] = true
		} else {
			m[task.ID] = false
		}
	}
	return n, m
}

func (zk *ZKStore) health(tasks types.TaskList) *types.Health {
	var (
		total     int64
		healthy   int64
		unhealthy int64
		unset     int64
	)

	for _, task := range tasks {
		switch task.Healthy {
		case types.TaskHealthy:
			healthy++
		case types.TaskUnHealthy:
			unhealthy++
		case types.TaskHealthyUnset:
			unset++
		}

		total++
	}

	return &types.Health{
		Total:     total,
		Healthy:   healthy,
		UnHealthy: unhealthy,
		UnSet:     unset,
	}
}

func (zk *ZKStore) version(tasks types.TaskList) []string {
	vers := make([]string, 0)

	for _, task := range tasks {
		if verExist(vers, task.Version) {
			continue
		}

		vers = append(vers, task.Version)
	}

	sort.Strings(vers)
	return vers
}

func verExist(vers []string, ver string) bool {
	for _, v := range vers {
		if v == ver {
			return true
		}
	}

	return false
}

func (zk *ZKStore) delTasks(p, id string) error {
	children, err := zk.list(path.Join(p, "tasks"))
	if err != nil {
		log.Errorf("get app %s children(tasks) error: %v", id, err)
		return err
	}

	for _, child := range children {
		if err := zk.del(path.Join(p, "tasks", child)); err != nil {
			log.Errorf("delete task %s got error: %v", child, err)
			return err
		}
	}

	if err := zk.del(path.Join(p, "tasks")); err != nil {
		log.Errorf("deleta znode %s/tasks got error: %v", p, err)
		return err
	}

	return nil
}

func (zk *ZKStore) delVersions(p, id string) error {
	children, err := zk.list(path.Join(p, "versions"))
	if err != nil {
		log.Errorf("get app %s children(versions) error: %v", id, err)
		return err
	}

	for _, child := range children {
		if err := zk.del(path.Join(p, "versions", child)); err != nil {
			log.Errorf("delete version %s got error: %v", child, err)
			return err
		}
	}

	if err := zk.del(path.Join(p, "versions")); err != nil {
		log.Errorf("deleta znode %s/versions got error: %v", p, err)
		return err
	}

	return nil
}
