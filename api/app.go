package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"
	"github.com/Dataman-Cloud/swan/utils/fields"
	"github.com/Dataman-Cloud/swan/utils/labels"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func (r *Server) createApp(w http.ResponseWriter, req *http.Request) {
	if err := checkForJSON(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var version types.Version
	if err := decode(req.Body, &version); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := version.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	version.ID = fmt.Sprintf("%d", time.Now().UTC().UnixNano())

	compose := req.Form.Get("compose")

	if compose == "" {
		compose = "default"
	}

	var (
		id        = fmt.Sprintf("%s.%s.%s.%s", version.Name, compose, version.RunAs, r.driver.ClusterName())
		count     = int(version.Instances)
		healthSet = version.HealthCheck != nil && !version.HealthCheck.IsEmpty()
		restart   = version.RestartPolicy
		retries   = 3
	)

	if restart != nil && restart.Retries > retries {
		retries = restart.Retries
	}

	app := &types.Application{
		ID:        id,
		Name:      version.Name,
		RunAs:     version.RunAs,
		Cluster:   r.driver.ClusterName(),
		OpStatus:  types.OpStatusCreating,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := r.db.CreateApp(app); err != nil {
		if strings.Contains(err.Error(), "app already exists") {
			http.Error(w, fmt.Sprintf("app %s has already exists", id), http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("create app error: %v", err), http.StatusInternalServerError)
		return
	}

	if err := r.db.CreateVersion(id, &version); err != nil {
		http.Error(w, fmt.Sprintf("create app version failed: %v", err), http.StatusInternalServerError)
		return
	}

	go func(appId string) {
		var err error

		// defer to mark op status
		defer func() {
			if err != nil {
				log.Errorf("launch app %s error: %v", appId, err)
				r.memoAppStatus(appId, types.OpStatusNoop, fmt.Sprintf("launch app error: %v", err), 0)
			} else {
				r.memoAppStatus(appId, types.OpStatusNoop, "", 0)
			}
		}()

		// prepare for all runtime tasks & db tasks
		log.Printf("Preparing to launch App %s with %d tasks", id, count)

		tasks := []*mesos.Task{}
		for i := 0; i < count; i++ {
			var (
				name = fmt.Sprintf("%d.%s", i, appId)
				id   = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
			)

			// runtime tasks
			cfg := types.NewTaskConfig(&version, i)
			t := mesos.NewTask(cfg, id, name)
			tasks = append(tasks, t)

			// save db tasks
			// TODO move db task creation to each runtime task logic
			task := &types.Task{
				ID:         id,
				Name:       name,
				Weight:     100,
				Status:     "pending",
				Healthy:    types.TaskHealthyUnset,
				Version:    version.ID,
				MaxRetries: retries,
				Created:    time.Now(),
				Updated:    time.Now(),
			}
			if healthSet {
				task.Healthy = types.TaskUnHealthy
			}

			log.Debugf("Create task %s in db", task.ID)
			if err = r.db.CreateTask(app.ID, task); err != nil {
				err = fmt.Errorf("create db task failed: %v", err)
				return
			}
		}

		err = r.driver.LaunchTasks(tasks)
		if err != nil {
			err = fmt.Errorf("launch tasks got error: %v", err)
			return
		}
	}(app.ID)

	writeJSON(w, http.StatusCreated, map[string]string{"Id": app.ID})
}

func (r *Server) listApps(w http.ResponseWriter, req *http.Request) {
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

func (r *Server) getApp(w http.ResponseWriter, req *http.Request) {
	// TODO(nmg): mux.Vars should be wrapped in context.
	id := mux.Vars(req)["app_id"]

	app, err := r.db.GetApp(id)
	if err != nil {
		if strings.Contains(err.Error(), "node does not exist") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, app)
}

func (r *Server) deleteApp(w http.ResponseWriter, req *http.Request) {
	var (
		appId = mux.Vars(req)["app_id"]
	)

	// get app
	_, err := r.db.GetApp(appId)
	if err != nil {
		if strings.Contains(err.Error(), "node does not exist") {
			http.Error(w, fmt.Sprintf("app %s not exists", appId), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get app tasks
	tasks, err := r.db.ListTasks(appId)
	if err != nil {
		http.Error(w, fmt.Sprintf("list tasks got error for delete app. %v", err), http.StatusInternalServerError)
		return
	}

	log.Debugf("app %s has %d tasks", appId, len(tasks))

	// get app versions
	versions, err := r.db.ListVersions(appId)
	if err != nil {
		http.Error(w, fmt.Sprintf("list tasks versions error for delete app. %v", err), http.StatusInternalServerError)
		return
	}

	log.Debugf("app %s has %d versions", appId, len(versions))

	// mark app op status
	if err := r.memoAppStatus(appId, types.OpStatusDeleting, "", 0); err != nil {
		http.Error(w, fmt.Sprintf("update app opstatus to deleting got error: %v", err), http.StatusInternalServerError)
		return
	}

	go func() {
		var err error

		// defer to mark op status
		defer func() {
			if err != nil {
				log.Errorf("delete app %s error: %v", appId, err)
				r.memoAppStatus(appId, types.OpStatusNoop, fmt.Sprintf("delete app error: %v", err), 0)
			}
		}()

		err = r.delApp(appId, tasks, versions)
	}()

	writeJSON(w, http.StatusNoContent, "")
}

func (r *Server) scaleApp(w http.ResponseWriter, req *http.Request) {
	appId := mux.Vars(req)["app_id"]

	app, err := r.db.GetApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if app.OpStatus != types.OpStatusNoop {
		http.Error(w, fmt.Sprintf("app status is %s, operation not allowed.", app.OpStatus), http.StatusLocked)
		return
	}

	var scale types.Scale
	if err := decode(req.Body, &scale); err != nil {
		http.Error(w, fmt.Sprintf("decode scale param error: %v", err), http.StatusBadRequest)
		return
	}

	tasks, err := r.db.ListTasks(appId)
	if err != nil {
		http.Error(w, fmt.Sprintf("list tasks got error for scale app. %v", err), http.StatusInternalServerError)
		return
	}

	var (
		current = len(tasks)
		goal    = scale.Instances
		ips     = scale.IPs // TODO(nmg): remove after automatic ipam
	)

	if goal < 0 {
		http.Error(w, "the goal count can't be negative", http.StatusBadRequest)
		return
	}

	if goal == current {
		writeJSON(w, http.StatusNotModified, "instances not changed")
		return
	}

	ver, err := r.db.GetVersion(app.ID, app.Version[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("get version got error for scale app. %v", err), http.StatusInternalServerError)
		return
	}

	newVer := ver
	newVer.ID = fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	newVer.Instances = int32(goal)
	newVer.IPs = ips

	if err := r.db.CreateVersion(appId, newVer); err != nil {
		http.Error(w, fmt.Sprintf("create app version failed: %v", err), http.StatusInternalServerError)
		return
	}

	if err := r.memoAppStatus(appId, types.OpStatusScaling, "", 0); err != nil {
		http.Error(w, fmt.Sprintf("update app opstatus to scaling got error: %v", err), http.StatusInternalServerError)
		return
	}

	if goal < current { // scale dwon
		go func() {
			var err error

			defer func() {
				if err != nil {
					log.Errorf("scale down app %s error: %v", appId, err)
					r.memoAppStatus(appId, types.OpStatusNoop, fmt.Sprintf("scale down app error: %v", err), 0)
				} else {
					r.memoAppStatus(appId, types.OpStatusNoop, "", 0)
				}
			}()

			types.TaskList(tasks).Reverse() // TODO
			var (
				killing = tasks[:current-goal]
				wg      sync.WaitGroup
				succeed int64
			)
			for _, task := range killing {
				wg.Add(1)

				go func(task *types.Task) {
					defer wg.Done()

					if err := r.driver.KillTask(task.ID, task.AgentId); err != nil {
						log.Errorf("Kill task %s got error: %v", task.ID, err)

						task.OpStatus = fmt.Sprintf("kill task error: %v", err)
						if err = r.db.UpdateTask(appId, task); err != nil {
							log.Errorf("update task %s got error: %v", task.Name, err)
						}

						return
					}

					if err := r.db.DeleteTask(task.ID); err != nil {
						log.Errorf("Delete task %s got error: %v", task.ID, err)

						task.OpStatus = fmt.Sprintf("delete task error: %v", err)
						if err = r.db.UpdateTask(appId, task); err != nil {
							log.Errorf("update task %s got error: %v", task.Name, err)
						}

						return
					}

					atomic.AddInt64(&succeed, 1)
				}(task)
			}
			wg.Wait()

			if int(succeed) != len(killing) {
				err = fmt.Errorf("%d tasks failed", len(killing)-int(succeed))
			}
		}()

		writeJSON(w, http.StatusAccepted, "accepted")
		return
	}

	// scale up
	version, err := r.db.GetVersion(appId, app.Version[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	net := version.Container.Docker.Network
	if net != "host" && net != "bridge" {
		if len(ips) < int(goal-current) {
			http.Error(w, fmt.Sprintf("IP number cannot be less than the instance number"), http.StatusBadRequest)
			return
		}
	}

	go func() {
		var err error

		// defer to mark op status
		defer func() {
			if err != nil {
				log.Errorf("scale up app %s error: %v", appId, err)
				r.memoAppStatus(appId, types.OpStatusNoop, fmt.Sprintf("scale up app error: %v", err), 0)
			} else {
				r.memoAppStatus(appId, types.OpStatusNoop, "", 0)
			}
		}()

		var (
			tasks     = []*mesos.Task{}
			healthSet = version.HealthCheck != nil && !version.HealthCheck.IsEmpty()
		)

		// prepare for all of runtime tasks & db tasks
		for i := current; i < goal; i++ {
			var (
				name    = fmt.Sprintf("%d.%s", i, appId)
				id      = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
				restart = version.RestartPolicy
				retries = 3
			)

			if restart != nil && restart.Retries > retries {
				retries = restart.Retries
			}

			// runtime tasks
			cfg := types.NewTaskConfig(version, i)
			t := mesos.NewTask(cfg, id, name)
			tasks = append(tasks, t)

			// db tasks
			task := &types.Task{
				ID:         id,
				Name:       name,
				Weight:     100,
				Status:     "pending",
				Healthy:    types.TaskHealthyUnset,
				Version:    version.ID,
				MaxRetries: retries,
				Created:    time.Now(),
				Updated:    time.Now(),
			}

			if healthSet {
				task.Healthy = types.TaskUnHealthy
			}

			if err = r.db.CreateTask(appId, task); err != nil {
				err = fmt.Errorf("create db task failed: %v", err)
				return
			}
		}

		err = r.driver.LaunchTasks(tasks)
		if err != nil {
			err = fmt.Errorf("launch tasks got error: %v", err)
			return
		}
	}()

	writeJSON(w, http.StatusAccepted, "accepted")
}

func (r *Server) updateApp(w http.ResponseWriter, req *http.Request) {
	appId := mux.Vars(req)["app_id"]

	app, err := r.db.GetApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if app.OpStatus != types.OpStatusNoop {
		http.Error(w, fmt.Sprintf("app status is %s, operation not allowed.", app.OpStatus), http.StatusLocked)
		return
	}

	newVer := new(types.Version)
	if err := decode(req.Body, newVer); err != nil {
		http.Error(w, fmt.Sprintf("decode update version got error: %v", err), http.StatusBadRequest)
		return
	}

	if err := newVer.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newVer.ID = fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	if err := r.db.CreateVersion(appId, newVer); err != nil {
		http.Error(w, fmt.Sprintf("create app version failed: %v", err), http.StatusInternalServerError)
		return
	}

	tasks, err := r.db.ListTasks(app.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("list tasks got error for update app. %v", err), http.StatusInternalServerError)
		return
	}

	if err := r.memoAppStatus(appId, types.OpStatusUpdating, "", 0); err != nil {
		http.Error(w, fmt.Sprintf("update app opstatus to rolling-update got error: %v", err), http.StatusInternalServerError)
		return
	}

	var (
		delay     = float64(1)
		onfailure = "stop"
	)

	update := newVer.UpdatePolicy
	if update != nil {
		delay = update.Delay
		onfailure = update.OnFailure
	}

	types.TaskList(tasks).Sort()
	pending := tasks

	go func() {
		defer func() {
			r.memoAppStatus(appId, types.OpStatusNoop, "", 0)
		}()

		var (
			progress  = 0
			healthSet = newVer.HealthCheck != nil && !newVer.HealthCheck.IsEmpty()
		)

		for i, t := range pending {
			progress++
			r.memoAppStatus(appId, types.OpStatusUpdating, "", progress) // TODO should quit if error occured here.

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

			cfg := types.NewTaskConfig(newVer, i)

			var (
				name    = t.Name
				id      = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
				restart = newVer.RestartPolicy
				retries = 3
			)

			if restart != nil && restart.Retries > retries {
				retries = restart.Retries
			}

			task := &types.Task{
				ID:         id,
				Name:       name,
				Weight:     100,
				Status:     "pending",
				Healthy:    types.TaskHealthyUnset,
				Version:    newVer.ID,
				MaxRetries: retries,
				Created:    t.Created,
				Updated:    time.Now(),
			}

			if healthSet {
				task.Healthy = types.TaskUnHealthy
			}

			if err := r.db.CreateTask(appId, task); err != nil {
				log.Errorf("create task failed: %s", err)
				return
			}

			m := mesos.NewTask(cfg, task.ID, task.Name)

			tasks := []*mesos.Task{m}

			if launchErr := r.driver.LaunchTasks(tasks); launchErr != nil {
				log.Errorf("launch task %s got error: %v", id, launchErr)

				task.Status = "Failed"
				task.ErrMsg = launchErr.Error()

				if err = r.db.UpdateTask(appId, task); err != nil {
					log.Errorf("update task %s got error: %v", id, err)
				}

				if onfailure == types.UpdateStop {
					return
				}
			}

			// notify proxy

			time.Sleep(time.Duration(delay) * time.Second)
		}
	}()

	writeJSON(w, http.StatusAccepted, "accepted")
}

func (s *Server) stopApp(w http.ResponseWriter, req *http.Request) {
	appId := mux.Vars(req)["app_id"]

	app, err := s.db.GetApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if app.OpStatus != types.OpStatusNoop {
		http.Error(w, fmt.Sprintf("app status is %s, operation not allowed.", app.OpStatus), http.StatusLocked)
		return
	}

	tasks, err := s.db.ListTasks(app.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("list tasks got error for stopping. %v", err), http.StatusInternalServerError)
		return
	}

	app.OpStatus = types.OpStatusStopping
	if err := s.db.UpdateApp(app); err != nil {
		http.Error(w, fmt.Sprintf("update app opstatus to stopping got error: %v", err), http.StatusInternalServerError)
		return
	}

	go func() {
		defer func() {
			app.OpStatus = types.OpStatusNoop
			if err := s.db.UpdateApp(app); err != nil {
				log.Errorf("update app opstatus from stopping to noop got error: %v", err)
			}
		}()
		var wg sync.WaitGroup
		for _, task := range tasks {
			wg.Add(1)
			go func(task *types.Task) {
				defer wg.Done()

				if err := s.delTask(appId, task); err != nil {
					log.Errorf("app %s stop task %s error: %v", appId, task.ID)
				}
			}(task)
		}

		wg.Wait()
	}()

	writeJSON(w, http.StatusAccepted, "accepted")
}

func (r *Server) canaryUpdate(w http.ResponseWriter, req *http.Request) {
	appId := mux.Vars(req)["app_id"]

	app, err := r.db.GetApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if app.OpStatus != types.OpStatusNoop {
		http.Error(w, fmt.Sprintf("app status is %s, operation not allowed.", app.OpStatus), http.StatusLocked)
		return
	}

	canary := new(types.CanaryUpdateBody)
	if err := decode(req.Body, canary); err != nil {
		http.Error(w, fmt.Sprintf("decode gray publish body got error: %v", err), http.StatusBadRequest)
		return
	}

	var (
		newVer    = canary.Version
		value     = canary.Value
		count     = canary.Instances
		onfailure = canary.OnFailure
		delay     = canary.Delay
	)

	if value == 0 {
		http.Error(w, "canary value must between (0, 1)", http.StatusInternalServerError)
		return
	}

	if count == 0 {
		count = 1
	}

	if newVer == nil {
		versions, err := r.db.ListVersions(app.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("list versions got error for canary update. %v", err), http.StatusInternalServerError)
			return
		}

		if len(versions) < 2 {
			http.Error(w, "canary version not specified and app has no new version", http.StatusInternalServerError)
			return
		}

		types.VersionList(versions).Reverse() // TODO

		newVer = versions[0]
	}

	if delay == 0 {
		delay = types.DefaultCanaryUpdateDelay
	}

	tasks, err := r.db.ListTasks(app.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("list tasks got error for canary update. %v", err), http.StatusInternalServerError)
		return
	}

	if err := r.memoAppStatus(appId, types.OpStatusUpdating, "", 0); err != nil {
		http.Error(w, fmt.Sprintf("update app opstatus to rolling-update got error: %v", err), http.StatusInternalServerError)
		return
	}

	types.TaskList(tasks).Sort() // TODO

	new := 0
	newTasks := make([]*types.Task, 0)
	for _, task := range tasks {
		if task.Version == newVer.ID {
			new++

			newTasks = append(newTasks, task)
		}
	}

	var (
		total = app.TaskCount
		goal  = new + count
	)

	if goal > total {
		goal = total
	}

	newWeight := utils.ComputeWeight(float64(goal), float64(total), value)

	pending := tasks[new:goal]

	go func() {
		defer func() {
			r.memoAppStatus(appId, types.OpStatusNoop, "", 0)
		}()

		healthSet := newVer.HealthCheck != nil && !newVer.HealthCheck.IsEmpty()

		for i, t := range pending {
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

			cfg := types.NewTaskConfig(newVer, i+new)

			var (
				name    = t.Name
				id      = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
				restart = newVer.RestartPolicy
				retries = 3
			)

			if restart != nil && restart.Retries > retries {
				retries = restart.Retries
			}

			task := &types.Task{
				ID:         id,
				Name:       name,
				Weight:     newWeight,
				Healthy:    types.TaskHealthyUnset,
				Version:    newVer.ID,
				MaxRetries: retries,
				Created:    t.Created,
				Updated:    time.Now(),
			}

			if healthSet {
				task.Healthy = types.TaskUnHealthy
			}

			if err := r.db.CreateTask(appId, task); err != nil {
				log.Errorf("create task failed: %s", err)
				return
			}

			m := mesos.NewTask(cfg, task.ID, task.Name)

			tasks := []*mesos.Task{m}

			if launchErr := r.driver.LaunchTasks(tasks); launchErr != nil {
				log.Errorf("launch task %s got error: %v", id, launchErr)

				task.Status = "Failed"
				task.ErrMsg = launchErr.Error()

				if err = r.db.UpdateTask(appId, task); err != nil {
					log.Errorf("update task %s got error: %v", id, err)
				}

				if onfailure == types.CanaryUpdateOnFailureStop {
					return
				}

			}

			// notify proxy

			time.Sleep(time.Duration(delay) * time.Second)
		}
	}()

	writeJSON(w, http.StatusAccepted, "accepted")
}

func (r *Server) rollback(w http.ResponseWriter, req *http.Request) {
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
		http.Error(w, fmt.Sprintf("app status is %s, operation not allowed.", app.OpStatus), http.StatusLocked)
		return
	}

	tasks, err := r.db.ListTasks(app.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("list tasks got error for rollback app. %v", err), http.StatusInternalServerError)
		return
	}

	versions, err := r.db.ListVersions(app.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("list versions got error for rollback app. %v", err), http.StatusInternalServerError)
		return
	}

	if err := r.memoAppStatus(appId, types.OpStatusRollback, "", 0); err != nil {
		http.Error(w, fmt.Sprintf("update app opstatus to rolling-back got error: %v", err), http.StatusInternalServerError)
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
		if len(versions) < 2 {
			http.Error(w, fmt.Sprintf("no more versions to rollback"), http.StatusInternalServerError)
			return
		}

		types.VersionList(versions).Sort()
		for idx, ver := range versions {
			if ver.ID == app.Version[0] {
				if (idx - 1) < 0 {
					http.Error(w, fmt.Sprintf("version error"), http.StatusInternalServerError)
					return
				}
				desired = versions[idx-1]
			}
		}
	}

	// TODO
	types.TaskList(tasks).Reverse()

	go func() {
		defer func() {
			r.memoAppStatus(appId, types.OpStatusNoop, "", 0)
		}()

		for i, t := range tasks {
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

			cfg := types.NewTaskConfig(desired, i)

			var (
				name    = t.Name
				id      = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
				restart = desired.RestartPolicy
				retries = 3
			)

			if restart != nil && restart.Retries > retries {
				retries = restart.Retries
			}

			task := &types.Task{
				ID:         id,
				Name:       name,
				Weight:     100,
				Status:     "updating",
				Version:    desired.ID,
				MaxRetries: retries,
				Created:    t.Created,
				Updated:    time.Now(),
			}

			if err := r.db.CreateTask(appId, task); err != nil {
				log.Errorf("create task failed: %s", err)
				return
			}

			m := mesos.NewTask(cfg, task.ID, task.Name)

			tasks := []*mesos.Task{m}

			if launchErr := r.driver.LaunchTasks(tasks); launchErr != nil {
				log.Errorf("launch task %s got error: %v", task.ID, launchErr)

				task.Status = "Failed"
				task.ErrMsg = fmt.Sprintf("launch task failed: %v", launchErr)

				if err = r.db.UpdateTask(appId, task); err != nil {
					log.Errorf("update task %s got error: %v", task.ID, err)
				}

				return
			}

			time.Sleep(2 * time.Second)

		}
	}()

	writeJSON(w, http.StatusAccepted, "accepted")
}

func (r *Server) updateWeights(w http.ResponseWriter, req *http.Request) {
	var body types.UpdateWeightsBody
	if err := decode(req.Body, &body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		appId   = mux.Vars(req)["app_id"]
		weights = body.Weights
	)

	app, err := r.db.GetApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tasks, err := r.db.ListTasks(app.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("list tasks got error for update weights. %v", err), http.StatusInternalServerError)
		return
	}

	for n, weight := range weights {
		for _, task := range tasks {
			if task.Index() == n {
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

func (r *Server) getTasks(w http.ResponseWriter, req *http.Request) {
	appId := mux.Vars(req)["app_id"]

	tasks, err := r.db.ListTasks(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (r *Server) getTask(w http.ResponseWriter, req *http.Request) {
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

func (r *Server) updateWeight(w http.ResponseWriter, req *http.Request) {
	var body types.UpdateWeightBody
	if err := decode(req.Body, &body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	task.Weight = body.Weight

	if err := r.db.UpdateTask(appId, task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// notify proxy

	writeJSON(w, http.StatusAccepted, "accepted")
}

func (r *Server) getVersions(w http.ResponseWriter, req *http.Request) {
	appId := mux.Vars(req)["app_id"]

	app, err := r.db.GetApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	versions, err := r.db.ListVersions(app.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("list versions got error for get versions. %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, versions)
}

func (r *Server) getVersion(w http.ResponseWriter, req *http.Request) {
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
func (r *Server) createVersion(w http.ResponseWriter, req *http.Request) {
	var (
		vars  = mux.Vars(req)
		appId = vars["app_id"]
	)

	if err := checkForJSON(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var version types.Version
	if err := decode(req.Body, &version); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := version.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

func (r *Server) deleteTask(w http.ResponseWriter, req *http.Request) {
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

func (r *Server) deleteTasks(w http.ResponseWriter, req *http.Request) {
	var (
		vars  = mux.Vars(req)
		appId = vars["app_id"]
	)

	app, err := r.db.GetApp(appId)
	if err != nil {
		if strings.Contains(err.Error(), "node does not exist") {
			http.Error(w, fmt.Sprintf("app %s not exists", appId), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tasks, err := r.db.ListTasks(app.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("list tasks got error for delete tasks. %v", err), http.StatusInternalServerError)
		return
	}

	for _, task := range tasks {
		go func(task *types.Task, appId string) {
			if err := r.driver.KillTask(task.ID, task.AgentId); err != nil {
				log.Errorf("Kill task %s got error: %v", task.ID, err)

				task.OpStatus = fmt.Sprintf("kill task error: %v", err)
				if err = r.db.UpdateTask(appId, task); err != nil {
					log.Errorf("update task %s got error: %v", task.Name, err)
				}

				return
			}

			if err := r.db.DeleteTask(task.ID); err != nil {
				log.Errorf("Kill task %s got error: %v", task.ID, err)

				task.OpStatus = fmt.Sprintf("delete task error: %v", err)
				if err = r.db.UpdateTask(appId, task); err != nil {
					log.Errorf("update task %s got error: %v", task.Name, err)
				}

				return
			}

		}(task, app.ID)
	}

	writeJSON(w, http.StatusNoContent, "")
}

func (r *Server) updateTask(w http.ResponseWriter, req *http.Request) {
	var (
		vars   = mux.Vars(req)
		appId  = vars["app_id"]
		taskId = vars["task_id"]
	)

	var version types.Version
	if err := decode(req.Body, &version); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := version.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	seq := strings.SplitN(t.Name, ".", 2)[0]
	idx, _ := strconv.Atoi(seq)
	cfg := types.NewTaskConfig(&version, idx)

	var (
		name    = t.Name
		id      = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
		restart = version.RestartPolicy
		retries = 3
	)

	if restart != nil && restart.Retries > retries {
		retries = restart.Retries
	}

	task := &types.Task{
		ID:         id,
		Name:       name,
		Weight:     100,
		Status:     "updating",
		Version:    version.ID,
		MaxRetries: retries,
		Created:    t.Created,
		Updated:    time.Now(),
	}

	if err := r.db.CreateTask(appId, task); err != nil {
		log.Errorf("create task failed: %s", err)
		return
	}

	m := mesos.NewTask(cfg, task.ID, task.Name)
	tasks := []*mesos.Task{m}

	if launchErr := r.driver.LaunchTasks(tasks); launchErr != nil {
		log.Errorf("launch task %s got error: %v", task.ID, launchErr)

		task.Status = "Failed"
		task.ErrMsg = fmt.Sprintf("launch task failed: %v", launchErr)

		if err = r.db.UpdateTask(appId, task); err != nil {
			log.Errorf("update task %s got error: %v", t.ID, err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusAccepted, "accepted")
}

func (r *Server) rollbackTask(w http.ResponseWriter, req *http.Request) {
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
		http.Error(w, fmt.Sprintf("app status is %s, operation not allowed.", app.OpStatus), http.StatusLocked)
		return
	}

	versions, err := r.db.ListVersions(app.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("list versions got error for rollback task. %v", err), http.StatusInternalServerError)
		return
	}

	if err := r.memoAppStatus(appId, types.OpStatusRollback, "", 0); err != nil {
		http.Error(w, fmt.Sprintf("update app opstatus to rolling-back got error: %v", err), http.StatusInternalServerError)
		return
	}

	defer func() {
		r.memoAppStatus(appId, types.OpStatusNoop, "", 0)
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
		if len(versions) < 2 {
			http.Error(w, fmt.Sprintf("no more versions to rollback"), http.StatusInternalServerError)
			return
		}

		// TODO
		types.VersionList(versions).Sort()
		for idx, ver := range versions {
			if ver.ID == app.Version[0] {
				if (idx - 1) < 0 {
					http.Error(w, fmt.Sprintf("version error"), http.StatusInternalServerError)
					return
				}
				desired = versions[idx-1]
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

	seq := strings.SplitN(t.Name, ".", 2)[0]
	idx, _ := strconv.Atoi(seq)
	cfg := types.NewTaskConfig(desired, idx)

	var (
		name    = t.Name
		id      = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
		restart = desired.RestartPolicy
		retries = 3
	)

	if restart != nil && restart.Retries > retries {
		retries = restart.Retries
	}

	task := &types.Task{
		ID:         id,
		Name:       name,
		Weight:     100,
		Status:     "updating",
		Version:    desired.ID,
		MaxRetries: retries,
		Created:    t.Created,
		Updated:    time.Now(),
	}

	if err := r.db.CreateTask(appId, task); err != nil {
		log.Errorf("create task failed: %s", err)
		return
	}

	m := mesos.NewTask(cfg, task.ID, task.Name)

	tasks := []*mesos.Task{m}

	if launchErr := r.driver.LaunchTasks(tasks); launchErr != nil {
		log.Errorf("launch task %s got error: %v", task.ID, launchErr)

		task.Status = "Failed"
		task.ErrMsg = fmt.Sprintf("launch task failed: %v", launchErr)

		if err = r.db.UpdateTask(appId, task); err != nil {
			log.Errorf("update task %s got error: %v", t.ID, err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusAccepted, "accepted")
}

// delApp actually remove App related runtime tasks & db objects
func (r *Server) delApp(appId string, tasks []*types.Task, versions []*types.Version) error {
	var (
		count   = len(tasks)
		succeed = int64(0)
		wg      sync.WaitGroup
	)

	// remove runtime tasks & db tasks firstly
	for _, task := range tasks {
		wg.Add(1)
		go func(task *types.Task) {
			defer wg.Done()

			// kill runtime task
			if err := r.driver.KillTask(task.ID, task.AgentId); err != nil {
				log.Errorf("Kill task %s got error: %v", task.ID, err)

				task.ErrMsg = fmt.Sprintf("kill task error: %v", err)
				if err = r.db.UpdateTask(appId, task); err != nil {
					log.Errorf("update task %s got error: %v", task.Name, err)
				}

				return
			}

			// remove db task
			if err := r.db.DeleteTask(task.ID); err != nil {
				log.Errorf("Delete task %s got error: %v", task.ID, err)

				task.ErrMsg = fmt.Sprintf("delete task error: %v", err)
				if err = r.db.UpdateTask(appId, task); err != nil {
					log.Errorf("update task %s got error: %v", task.Name, err)
				}
				return
			}

			atomic.AddInt64(&succeed, 1)
		}(task)
	}
	wg.Wait()

	if int(succeed) != count {
		return fmt.Errorf("%d tasks kill / removed failed", count-int(succeed))
	}

	// remove db versions
	for _, version := range versions {
		if err := r.db.DeleteVersion(appId, version.ID); err != nil {
			return fmt.Errorf("Delete version %s for app %s got error: %v", version.ID, appId, err)
		}
	}

	// remove db app
	if err := r.db.DeleteApp(appId); err != nil {
		log.Errorf("Delete app %s got error: %v", appId, err)
	}

	return nil
}

func (s *Server) delTask(appId string, task *types.Task) error {
	if err := s.driver.KillTask(task.ID, task.AgentId); err != nil {
		log.Errorf("Kill task %s got error: %v", task.ID, err)

		task.OpStatus = fmt.Sprintf("kill task error: %v", err)
		if err = s.db.UpdateTask(appId, task); err != nil {
			log.Errorf("update task %s got error: %v", task.Name, err)
		}

		return err
	}

	if err := s.db.DeleteTask(task.ID); err != nil {
		log.Errorf("Delete task %s got error: %v", task.ID, err)

		task.OpStatus = fmt.Sprintf("delete task error: %v", err)
		if err = s.db.UpdateTask(appId, task); err != nil {
			log.Errorf("update task %s got error: %v", task.Name, err)
		}
	}

	return nil
}

// short hands to memo update App.OpStatus & App.ErrMsg & App.Progress
// it's the caller responsibility to process the db error.
func (r *Server) memoAppStatus(appId, op, errmsg string, progress int) error {
	app, err := r.db.GetApp(appId)
	if err != nil {
		log.Errorf("memoAppStatus() get db app %s error: %v", appId, err)
		return err
	}

	var (
		prevOp  = app.OpStatus
		prevPrg = app.Progress
	)

	app.OpStatus = op
	app.Progress = progress
	app.ErrMsg = errmsg
	app.UpdatedAt = time.Now()

	if err := r.db.UpdateApp(app); err != nil {
		log.Errorf("memoAppStatus() update app db status from %s(%d) -> %s(%d) error: %v",
			prevOp, prevPrg, op, progress, err)
		return err
	}

	return nil
}
