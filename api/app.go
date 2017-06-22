package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"
	"github.com/Dataman-Cloud/swan/utils/fields"
	"github.com/Dataman-Cloud/swan/utils/labels"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func (r *Router) createApp(w http.ResponseWriter, req *http.Request) {
	if err := checkForJSON(req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var version types.Version
	if err := decode(req.Body, &version); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := version.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	compose := req.Form.Get("compose")

	if compose == "" {
		compose = "default"
	}

	var (
		spec  = &version
		id    = fmt.Sprintf("%s.%s.%s.%s", spec.Name, compose, spec.RunAs, r.driver.ClusterName())
		vid   = fmt.Sprintf("%d", time.Now().UTC().UnixNano())
		count = int(spec.Instances)
	)

	var alias string
	if spec.Proxy != nil {
		alias = spec.Proxy.Alias
	}

	app := &types.Application{
		ID:        id,
		Name:      spec.Name,
		Alias:     alias,
		RunAs:     spec.RunAs,
		Cluster:   r.driver.ClusterName(),
		Status:    "creating",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := r.db.CreateApp(app); err != nil {
		if strings.Contains(err.Error(), "app already exists") {
			http.Error(w, fmt.Sprintf("app %s has already exists", id), http.StatusConflict)
		} else {

			http.Error(w, fmt.Sprintf("create app error: %v", err), http.StatusInternalServerError)
		}
		return
	}

	version.ID = vid

	if err := r.db.CreateVersion(id, &version); err != nil {
		http.Error(w, fmt.Sprintf("create app version failed: %v", err), http.StatusInternalServerError)
		return
	}

	//var (
	//	step      = spec.DeployPolicy.Step
	//	onfailure = spec.DeployPolicy.OnFailure
	//)

	//if step > types.MaxDeployStep {
	//	step = types.MaxDeployStep
	//}

	go func() {
		defer func() {
			app.OpStatus = types.OpStatusNoop
			if err := r.db.UpdateApp(app); err != nil {
				log.Errorf("update app op-status got error: %v", err)
			}
		}()

		for i := 0; i < count; i++ {
			//for j := i * step; j < i*step; j++ {

			//}
			var (
				name = fmt.Sprintf("%d.%s", i, app.ID)
				id   = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
			)

			cfg := types.NewTaskConfig(spec)

			if cfg.Network != "host" && cfg.Network != "bridge" {
				cfg.Parameters = append(cfg.Parameters, &types.Parameter{
					Key:   "ip",
					Value: spec.IPs[i],
				})

				cfg.IP = spec.IPs[i]
			}

			task := &types.Task{
				ID:      id,
				Name:    name,
				Weight:  100,
				Status:  "creating",
				Healthy: types.TaskHealthyUnset,
				Version: vid,
				Created: time.Now(),
				Updated: time.Now(),
			}

			if err := r.db.CreateTask(app.ID, task); err != nil {
				log.Errorf("create task failed: %s", err)
				break
			}

			t := mesos.NewTask(
				cfg,
				id,
				name,
			)

			if err := r.driver.LaunchTask(t); err != nil {
				log.Errorf("launch task %s got error: %v", id, err)

				task.Status = "Failed"
				task.ErrMsg = err.Error()

				if err = r.db.UpdateTask(app.ID, task); err != nil {
					log.Errorf("update task %s status got error: %v", id, err)
				}

				break
			}
		}

		return
	}()

	writeJSON(w, http.StatusCreated, map[string]string{"Id": app.ID})
}

func (r *Router) listApps(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filters := types.AppFilterOptions{}

	labelsFilter := req.Form.Get("labels")
	if labelsFilter != "" {
		selector, err := labels.Parse(labelsFilter)
		if err != nil {
			http.Error(w, fmt.Sprintf("parse labels %s failed: %v", selector, err), http.StatusBadRequest)
			return
		}
		filters.LabelsSelector = selector
	}

	fieldsFilter := req.Form.Get("fields")
	if fieldsFilter != "" {
		selector, err := fields.ParseSelector(fieldsFilter)
		if err != nil {
			http.Error(w, fmt.Sprintf("parse labels %s failed: %v", selector, err), http.StatusBadRequest)
			return
		}

		filters.FieldsSelector = selector
	}

	rets, err := r.db.ListApps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apps := make([]*types.Application, 0)
	for _, app := range rets {
		ver, err := r.db.GetVersion(app.ID, app.Version[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !filterByLabelsSelectors(filters.LabelsSelector, ver.Labels) {
			continue
		}

		if !filterByFieldsSelectors(filters.FieldsSelector, ver) {
			continue
		}

		apps = append(apps, app)
	}

	writeJSON(w, http.StatusOK, apps)
}

func (r *Router) getApp(w http.ResponseWriter, req *http.Request) {
	// TODO(nmg): mux.Vars should be wrapped in context.
	id := mux.Vars(req)["app_id"]

	app, err := r.db.GetApp(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, app)
}

func (r *Router) deleteApp(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["app_id"]

	app, err := r.db.GetApp(id)
	if err != nil {
		if strings.Contains(err.Error(), "node does not exist") {
			http.Error(w, fmt.Sprintf("app %s not exists", id), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	var (
		errch chan error
		wg    sync.WaitGroup
	)

	for _, task := range app.Tasks {
		wg.Add(1)
		go func(errch chan error) {
			if err := r.driver.KillTask(task.ID, task.AgentId); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				errch <- err
				return
			}

			if err := r.db.DeleteTask(task.ID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				errch <- err
				return
			}

			wg.Done()
		}(errch)
	}

	select {
	case err := <-errch:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		wg.Wait()
	}

	if err := r.db.DeleteApp(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusNoContent, "")
}

func (r *Router) scale(w http.ResponseWriter, req *http.Request) {
	appId := mux.Vars(req)["app_id"]

	app, err := r.db.GetApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if app.OpStatus != types.OpStatusNoop {
		http.Error(w, fmt.Sprintf("app status is %s, operation not allowed.", app.OpStatus), http.StatusMethodNotAllowed)
		return
	}

	var param types.ScaleParam
	if err := decode(req.Body, &param); err != nil {
		http.Error(w, fmt.Sprintf("decode scale param error: %v", err), http.StatusInternalServerError)
		return
	}

	var (
		current = len(app.Tasks)
		goal    = param.Instances
		ips     = param.IPs
	)

	if goal < 0 {
		goal = -goal
	}

	if goal == current {
		writeJSON(w, http.StatusNotModified, "instances not changed")
		return
	}

	app.OpStatus = types.OpStatusScaling

	if err := r.db.UpdateApp(app); err != nil {
		http.Error(w, fmt.Sprintf("updating app opstatus to scaling got error: %v", err.Error), http.StatusInternalServerError)
		return
	}

	if goal < current { // scale dwon
		go func() {
			defer func() {
				app.OpStatus = types.OpStatusNoop
				if err := r.db.UpdateApp(app); err != nil {
					log.Errorf("updating app status from scaling to noop got error: %v", err)
				}
			}()

			for i := 1; i < current-goal+1; i++ {
				tid := fmt.Sprintf("%d.%s", current-i, appId)

				t, err := r.db.GetTask(appId, tid)
				if err != nil {
					break
				}

				if err := r.driver.KillTask(t.ID, t.AgentId); err != nil {
					t.Status = "delete failed"
					t.ErrMsg = err.Error()

					if err = r.db.UpdateTask(appId, t); err != nil {
						log.Errorf("update task %s got error: %v", t.Name, err)
					}

					break
				}

				if err := r.db.DeleteTask(t.ID); err != nil {
					log.Errorf("delete task %s got error: %v", t.Name, err)
					break
				}
			}
		}()

		writeJSON(w, http.StatusAccepted, "accepted")
		return
	}

	// scale up
	spec, err := r.db.GetVersion(app.ID, app.Version[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	net := spec.Container.Docker.Network
	if net != "host" && net != "bridge" {
		if len(ips) < int(goal-current) {
			http.Error(w, fmt.Sprintf("IP number cannot be less than the instance number"), http.StatusBadRequest)
			return
		}
	}

	go func() {
		defer func() {
			app.OpStatus = types.OpStatusNoop
			if err := r.db.UpdateApp(app); err != nil {
				log.Errorf("updating app status from scaling to noop got error: %v", err)
			}
		}()

		for i := current; i < goal; i++ {
			var (
				name = fmt.Sprintf("%d.%s", i, appId)
				id   = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
				ver  = spec.ID
			)

			cfg := types.NewTaskConfig(spec)

			if cfg.Network != "host" && cfg.Network != "bridge" {
				cfg.Parameters = append(cfg.Parameters, &types.Parameter{
					Key:   "ip",
					Value: ips[i],
				})

				cfg.IP = spec.IPs[i]
			}

			task := &types.Task{
				ID:      id,
				Name:    name,
				Weight:  100,
				Status:  "creating",
				Version: ver,
				Created: time.Now(),
				Updated: time.Now(),
			}

			if err := r.db.CreateTask(appId, task); err != nil {
				log.Errorf("create task failed: %s", err)
				break
			}

			t := mesos.NewTask(
				cfg,
				id,
				name,
			)

			if err := r.driver.LaunchTask(t); err != nil {
				log.Errorf("launch task %s got error: %v", id, err)

				task.Status = "Failed"
				task.ErrMsg = err.Error()

				if err = r.db.UpdateTask(appId, task); err != nil {
					log.Errorf("update task %s got error: %v", id, err)
				}

				return
			}
		}
	}()

	writeJSON(w, http.StatusAccepted, "accepted")
}

func (r *Router) updateApp(w http.ResponseWriter, req *http.Request) {
	var version types.Version
	if err := decode(req.Body, &version); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := version.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	appId := mux.Vars(req)["app_id"]

	version.ID = fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	if err := r.db.CreateVersion(appId, &version); err != nil {
		http.Error(w, fmt.Sprintf("create app version failed: %v", err), http.StatusInternalServerError)
		return
	}

	app, err := r.db.GetApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if app.OpStatus != types.OpStatusNoop {
		http.Error(w, fmt.Sprintf("app status is %s, operation not allowed.", app.OpStatus), http.StatusMethodNotAllowed)
		return
	}

	app.OpStatus = types.OpStatusUpdating
	app.UpdatingVersion = version.ID

	if err := r.db.UpdateApp(app); err != nil {
		http.Error(w, fmt.Sprintf("updating app opstatus to rolling-update got error: %v", err.Error), http.StatusInternalServerError)
		return
	}

	sort.Sort(types.TaskSorter(app.Tasks))

	spec := &version

	go func() {
		defer func() {
			app.OpStatus = types.OpStatusNoop
			if err := r.db.UpdateApp(app); err != nil {
				log.Errorf("updating app status from updating to noop got error: %v", err)
			}
		}()

		for _, t := range app.Tasks {
			if err := r.driver.KillTask(t.ID, t.AgentId); err != nil {
				t.Status = "Failed"
				t.ErrMsg = fmt.Sprintf("kill task for updating :%v", err)

				if err = r.db.UpdateTask(app.ID, t); err != nil {
					log.Errorf("update task %s got error: %v", t.ID, err)
				}

				return
			}

			if err := r.db.DeleteTask(t.ID); err != nil {
				log.Errorf("delete task %s got error: %v", t.ID, err)
				return
			}

			cfg := types.NewTaskConfig(spec)

			var (
				name = t.Name
				id   = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
			)

			task := &types.Task{
				ID:      id,
				Name:    name,
				Weight:  100,
				Status:  "updating",
				Version: spec.ID,
				Created: t.Created,
				Updated: time.Now(),
			}

			if err := r.db.CreateTask(appId, task); err != nil {
				log.Errorf("create task failed: %s", err)
				return
			}

			mtask := mesos.NewTask(cfg, task.ID, task.Name)

			if err := r.driver.LaunchTask(mtask); err != nil {
				log.Errorf("launch task %s got error: %v", t.ID, err)

				task.Status = "Failed"
				task.ErrMsg = fmt.Sprintf("launch task failed: %v", err)

				if err = r.db.UpdateTask(appId, task); err != nil {
					log.Errorf("update task %s got error: %v", t.ID, err)
				}

				return
			}

			time.Sleep(5 * time.Second)
		}
	}()

	writeJSON(w, http.StatusAccepted, "accepted")
}

func (r *Router) cancelUpdate(w http.ResponseWriter, req *http.Request) {
}

func (r *Router) rollback(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	appId := mux.Vars(req)["app_id"]

	app, err := r.db.GetApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if app.OpStatus != types.OpStatusNoop {
		http.Error(w, fmt.Sprintf("app status is %s, operation not allowed.", app.OpStatus), http.StatusMethodNotAllowed)
		return
	}

	app.OpStatus = types.OpStatusRollback

	if err := r.db.UpdateApp(app); err != nil {
		http.Error(w, fmt.Sprintf("updating app opstatus to rolling-back got error: %v", err.Error), http.StatusInternalServerError)
		return
	}

	verId := req.Form.Get("version")

	var desired *types.Version

	if verId != "" {
		ver, err := r.db.GetVersion(appId, verId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		desired = ver
	}

	if verId == "" {
		if len(app.Versions) < 2 {
			http.Error(w, fmt.Sprintf("no more versions to rollback"), http.StatusInternalServerError)
			return
		}

		vers := app.Versions.Sort()
		for idx, ver := range vers {
			if ver.ID == app.Version[0] {
				if (idx - 1) < 0 {
					http.Error(w, fmt.Sprintf("version error"), http.StatusInternalServerError)
					return
				}
				desired = vers[idx-1]
			}
		}
	}

	tasks := app.Tasks.Reverse()

	go func() {
		defer func() {
			app.OpStatus = types.OpStatusNoop
			if err := r.db.UpdateApp(app); err != nil {
				log.Errorf("updating app status from rollback to noop got error: %v", err)
			}
		}()

		for _, t := range tasks {
			if err := r.driver.KillTask(t.ID, t.AgentId); err != nil {
				t.Status = "Failed"
				t.ErrMsg = fmt.Sprintf("kill task for rollback :%v", err)

				if err = r.db.UpdateTask(appId, t); err != nil {
					log.Errorf("update task %s got error: %v", t.ID, err)
				}

				return
			}

			if err := r.db.DeleteTask(t.ID); err != nil {
				log.Errorf("delete task %s got error: %v", t.ID, err)
				return
			}

			cfg := types.NewTaskConfig(desired)

			var (
				name = t.Name
				id   = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
			)

			task := &types.Task{
				ID:      id,
				Name:    name,
				Weight:  100,
				Status:  "updating",
				Version: desired.ID,
				Created: t.Created,
				Updated: time.Now(),
			}

			if err := r.db.CreateTask(appId, task); err != nil {
				log.Errorf("create task failed: %s", err)
				return
			}

			mtask := mesos.NewTask(cfg, task.ID, task.Name)

			if err := r.driver.LaunchTask(mtask); err != nil {
				log.Errorf("launch task %s got error: %v", t.ID, err)

				task.Status = "Failed"
				task.ErrMsg = fmt.Sprintf("launch task failed: %v", err)

				if err = r.db.UpdateTask(appId, task); err != nil {
					log.Errorf("update task %s got error: %v", t.ID, err)
				}

				return
			}

			time.Sleep(5 * time.Second)

		}
	}()

	writeJSON(w, http.StatusAccepted, "accepted")
}

func (r *Router) stopScale(w http.ResponseWriter, req *http.Request) {
}

func (r *Router) stopUpdate(w http.ResponseWriter, req *http.Request) {
}

func (r *Router) updateWeights(w http.ResponseWriter, req *http.Request) {
	var param types.UpdateWeightsParam
	if err := decode(req.Body, &param); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var (
		appId   = mux.Vars(req)["app_id"]
		weights = param.Weights
	)

	app, err := r.db.GetApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for n, weight := range weights {
		for _, task := range app.Tasks {
			if task.Name == n {
				task.Weight = weight

				if err := r.db.UpdateTask(appId, task); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	}

	// notify proxy

}

func (r *Router) getTasks(w http.ResponseWriter, req *http.Request) {
	appId := mux.Vars(req)["app_id"]

	app, err := r.db.GetApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, app.Tasks)
}

func (r *Router) getTask(w http.ResponseWriter, req *http.Request) {
	var (
		vars   = mux.Vars(req)
		appId  = vars["app_id"]
		taskId = vars["task_id"]
	)

	task, err := r.db.GetTask(appId, taskId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (r *Router) updateWeight(w http.ResponseWriter, req *http.Request) {
	var weight types.UpdateWeightParam
	if err := decode(req.Body, &weight); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var (
		vars   = mux.Vars(req)
		appId  = vars["app_id"]
		taskId = vars["task_id"]
	)

	task, err := r.db.GetTask(appId, taskId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	task.Weight = weight.Weight

	if err := r.db.UpdateTask(appId, task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// notify proxy

	writeJSON(w, http.StatusAccepted, "accepted")
}

func (r *Router) getVersions(w http.ResponseWriter, req *http.Request) {
	appId := mux.Vars(req)["app_id"]

	app, err := r.db.GetApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, app.Versions)

}

func (r *Router) getVersion(w http.ResponseWriter, req *http.Request) {
	var (
		vars  = mux.Vars(req)
		appId = vars["app_id"]
		verId = vars["version_id"]
	)

	version, err := r.db.GetVersion(appId, verId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, version)
}

// TODO(nmg): named version should be supported.
func (r *Router) createVersion(w http.ResponseWriter, req *http.Request) {
	var (
		vars  = mux.Vars(req)
		appId = vars["app_id"]
	)

	if err := checkForJSON(req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var version types.Version
	if err := decode(req.Body, &version); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := version.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	version.ID = fmt.Sprintf("%d", time.Now().UTC().UnixNano())

	if err := r.db.CreateVersion(appId, &version); err != nil {
		http.Error(w, fmt.Sprintf("create app version failed: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"Id": version.ID})
}

func filterByLabelsSelectors(labelsSelector labels.Selector, appLabels map[string]string) bool {
	if labelsSelector == nil {
		return true
	}

	return labelsSelector.Matches(labels.Set(appLabels))
}

func filterByFieldsSelectors(fieldSelector fields.Selector, ver *types.Version) bool {
	if fieldSelector == nil {
		return true
	}

	// TODO(upccup): there maybe exist better way to got a field/value map
	fieldMap := make(map[string]string)
	fieldMap["runAs"] = ver.RunAs
	return fieldSelector.Matches(fields.Set(fieldMap))
}

func (r *Router) deleteTask(w http.ResponseWriter, req *http.Request) {
	var (
		vars   = mux.Vars(req)
		appId  = vars["app_id"]
		taskId = vars["task_id"]
	)

	task, err := r.db.GetTask(appId, taskId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := r.driver.KillTask(task.ID, task.AgentId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := r.db.DeleteTask(task.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusNoContent, "")
}

func (r *Router) updateTask(w http.ResponseWriter, req *http.Request) {
	var (
		vars   = mux.Vars(req)
		appId  = vars["app_id"]
		taskId = vars["task_id"]
	)

	var version types.Version
	if err := decode(req.Body, &version); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := version.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	version.ID = fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	if err := r.db.CreateVersion(appId, &version); err != nil {
		http.Error(w, fmt.Sprintf("create app version failed: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := r.db.GetTask(appId, taskId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := r.driver.KillTask(t.ID, t.AgentId); err != nil {
		t.Status = "Failed"
		t.ErrMsg = fmt.Sprintf("kill task for updating :%v", err)

		if err = r.db.UpdateTask(appId, t); err != nil {
			log.Errorf("update task %s got error: %v", t.ID, err)
		}

		return
	}

	if err := r.db.DeleteTask(t.ID); err != nil {
		log.Errorf("delete task %s got error: %v", t.ID, err)
		return
	}

	cfg := types.NewTaskConfig(&version)

	var (
		name = t.Name
		id   = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
	)

	task := &types.Task{
		ID:      id,
		Name:    name,
		Weight:  100,
		Status:  "updating",
		Version: version.ID,
		Created: t.Created,
		Updated: time.Now(),
	}

	if err := r.db.CreateTask(appId, task); err != nil {
		log.Errorf("create task failed: %s", err)
		return
	}

	mtask := mesos.NewTask(cfg, task.ID, task.Name)

	if err := r.driver.LaunchTask(mtask); err != nil {
		log.Errorf("launch task %s got error: %v", t.ID, err)

		task.Status = "Failed"
		task.ErrMsg = fmt.Sprintf("launch task failed: %v", err)

		if err = r.db.UpdateTask(appId, task); err != nil {
			log.Errorf("update task %s got error: %v", t.ID, err)
		}

		return
	}

	writeJSON(w, http.StatusAccepted, "accepted")
}

func (r *Router) rollbackTask(w http.ResponseWriter, req *http.Request) {
	var (
		vars   = mux.Vars(req)
		appId  = vars["app_id"]
		taskId = vars["task_id"]
	)

	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app, err := r.db.GetApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if app.OpStatus != types.OpStatusNoop {
		http.Error(w, fmt.Sprintf("app status is %s, operation not allowed.", app.OpStatus), http.StatusMethodNotAllowed)
		return
	}

	app.OpStatus = types.OpStatusRollback

	if err := r.db.UpdateApp(app); err != nil {
		http.Error(w, fmt.Sprintf("updating app opstatus to rolling-back got error: %v", err.Error), http.StatusInternalServerError)
		return
	}

	defer func() {
		app.OpStatus = types.OpStatusNoop
		if err := r.db.UpdateApp(app); err != nil {
			log.Errorf("updating app status from rollback to noop got error: %v", err)
		}
	}()

	verId := req.Form.Get("version")

	var desired *types.Version

	if verId != "" {
		ver, err := r.db.GetVersion(appId, verId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		desired = ver
	}

	if verId == "" {
		if len(app.Versions) < 2 {
			http.Error(w, fmt.Sprintf("no more versions to rollback"), http.StatusInternalServerError)
			return
		}

		vers := app.Versions.Sort()
		for idx, ver := range vers {
			if ver.ID == app.Version[0] {
				if (idx - 1) < 0 {
					http.Error(w, fmt.Sprintf("version error"), http.StatusInternalServerError)
					return
				}
				desired = vers[idx-1]
			}
		}
	}

	t, err := r.db.GetTask(appId, taskId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := r.driver.KillTask(t.ID, t.AgentId); err != nil {
		t.Status = "Failed"
		t.ErrMsg = fmt.Sprintf("kill task for rollback :%v", err)

		if err = r.db.UpdateTask(appId, t); err != nil {
			log.Errorf("update task %s got error: %v", t.ID, err)
		}

		return
	}

	if err := r.db.DeleteTask(t.ID); err != nil {
		log.Errorf("delete task %s got error: %v", t.ID, err)
		return
	}

	cfg := types.NewTaskConfig(desired)

	var (
		name = t.Name
		id   = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
	)

	task := &types.Task{
		ID:      id,
		Name:    name,
		Weight:  100,
		Status:  "updating",
		Version: desired.ID,
		Created: t.Created,
		Updated: time.Now(),
	}

	if err := r.db.CreateTask(appId, task); err != nil {
		log.Errorf("create task failed: %s", err)
		return
	}

	mtask := mesos.NewTask(cfg, task.ID, task.Name)

	if err := r.driver.LaunchTask(mtask); err != nil {
		log.Errorf("launch task %s got error: %v", t.ID, err)

		task.Status = "Failed"
		task.ErrMsg = fmt.Sprintf("launch task failed: %v", err)

		if err = r.db.UpdateTask(appId, task); err != nil {
			log.Errorf("update task %s got error: %v", t.ID, err)
		}

		return
	}

	writeJSON(w, http.StatusAccepted, "accepted")
}
