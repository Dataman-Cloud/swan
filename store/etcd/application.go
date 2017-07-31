package etcd

import (
	"path"
	"sort"

	log "github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *EtcdStore) CreateApp(app *types.Application) error {
	var (
		p    = path.Join(keyApp, app.ID)
		pval = path.Join(p, "value")
	)

	for _, sub := range []string{keyTasks, keyVersions} {
		subp := path.Join(p, sub)
		if err := s.ensureDir(subp); err != nil {
			return err
		}
	}

	data, err := s.get(pval)
	if err != nil {
		if !isEtcdKeyNotFound(err) {
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

	return s.create(pval, bs)
}

func (s *EtcdStore) UpdateApp(app *types.Application) error {
	var (
		p    = path.Join(keyApp, app.ID)
		pval = path.Join(p, "value")
	)

	bs, err := encode(app)
	if err != nil {
		return err
	}

	return s.update(pval, bs)
}

func (s *EtcdStore) GetApp(id string) (*types.Application, error) {
	var (
		p    = path.Join(keyApp, id)
		pval = path.Join(p, "value")
	)

	data, err := s.get(pval)
	if err != nil {
		log.Errorf("find app %s got error: %v", id, err)
		return nil, err
	}

	var app types.Application
	if err := decode(data, &app); err != nil {
		return nil, err
	}

	tasks, err := s.tasks(p, id)
	if err != nil {
		log.Errorf("get app %s tasks got error: %v", id, err)
		return nil, err
	}

	app.TaskCount = len(tasks)
	app.Status = s.status(tasks)
	app.TasksStatus = s.tasksStatus(tasks)
	app.Version = s.version(tasks)
	app.Health = s.health(tasks)

	versions, err := s.versions(p, id)
	if err != nil {
		log.Errorf("get app %s versions got error: %v", id, err)
		return nil, err
	}

	app.VersionCount = len(versions)

	if len(app.Version) == 0 {
		types.VersionList(versions).Reverse()
		app.Version = append(app.Version, versions[0].ID)
	}

	return &app, nil
}

func (s *EtcdStore) ListApps() ([]*types.Application, error) {
	nodes, err := s.list(keyApp)
	if err != nil {
		log.Errorln("etcd ListApps error:", err)
		return nil, err
	}

	apps := make([]*types.Application, 0)
	for id := range nodes {
		app, err := s.GetApp(id)
		if err != nil {
			log.Errorf("%v", err)
			continue
		}
		apps = append(apps, app)
	}

	return apps, nil
}

func (s *EtcdStore) DeleteApp(id string) error {
	p := path.Join(keyApp, id)
	return s.delDir(p, true)
}

func (s *EtcdStore) tasks(p, id string) (types.TaskList, error) {
	children, err := s.list(path.Join(p, keyTasks))
	if err != nil {
		log.Errorf("get app %s children(tasks) error: %v", id, err)
		return nil, err
	}

	tasks := make([]*types.Task, 0)
	for child := range children {
		p := path.Join(keyApp, id, keyTasks, child)
		data, err := s.get(p)
		if err != nil {
			log.Errorf("get %s got error: %v", p, err)
			return nil, err
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

func (s *EtcdStore) versions(p, id string) (types.VersionList, error) {
	children, err := s.list(path.Join(p, keyVersions))
	if err != nil {
		log.Errorf("get app %s children(versions) error: %v", id, err)
		return nil, err
	}

	versions := make([]*types.Version, 0)
	for child := range children {
		p := path.Join(keyApp, id, keyVersions, child)
		data, err := s.get(p)
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

func (s *EtcdStore) status(tasks types.TaskList) string {
	for _, task := range tasks {
		if task.Status == "TASK_RUNNING" {
			return "available"
		}
	}

	return "unavailable"
}

func (s *EtcdStore) tasksStatus(tasks types.TaskList) map[string]int {
	ret := make(map[string]int)
	for _, task := range tasks {
		ret[task.Status]++
	}
	return ret
}

func (s *EtcdStore) health(tasks types.TaskList) *types.Health {
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

func (s *EtcdStore) version(tasks types.TaskList) []string {
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
