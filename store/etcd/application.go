package etcd

import (
	"fmt"
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

	tasks, err := s.ListTasks(id)
	if err != nil {
		log.Errorf("get app %s tasks got error: %v", id, err)
		return nil, err
	}

	app.TaskCount = len(tasks)
	app.Status = s.status(tasks)
	app.TasksStatus = s.tasksStatus(tasks)
	app.Version = s.version(tasks)
	app.Health = s.health(tasks)

	versions, err := s.ListVersions(id)
	if err != nil {
		log.Errorf("get app %s versions got error: %v", id, err)
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("app %s versions temporary not ready", id)
	}

	app.VersionCount = len(versions)

	if len(app.Version) == 0 {
		types.VersionList(versions).Reverse()
		app.Version = append(app.Version, versions[0].ID)
	}

	return &app, nil
}

func (s *EtcdStore) GetAppOpStatus(appId string) (string, error) {
	var (
		p    = path.Join(keyApp, appId)
		pval = path.Join(p, "value")
	)

	data, err := s.get(pval)
	if err != nil {
		log.Errorf("find app %s got error: %v", appId, err)
		return "", err
	}

	var app types.Application
	if err := decode(data, &app); err != nil {
		return "", err
	}

	return app.OpStatus, nil
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
