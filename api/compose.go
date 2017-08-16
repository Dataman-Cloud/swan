package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func (r *Server) runCompose(w http.ResponseWriter, req *http.Request) {
	var err error

	if err = checkForJSON(req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var cmp types.Compose
	if err = decode(req.Body, &cmp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// check conflict
	if cmp, _ := r.db.GetCompose(cmp.Name); cmp != nil {
		http.Error(w, fmt.Sprintf("compose %s already exists", cmp.Name), http.StatusConflict)
		return
	}

	// convert raw yaml to service group
	if cmp.RequireConvert() {
		if cmp.ServiceGroup, err = cmp.ToServiceGroup(); err != nil {
			http.Error(w, fmt.Sprintf("yaml convert error: %v", err), http.StatusBadRequest)
			return
		}
	}

	// verify
	if err := cmp.Valid(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ensure all settings could be converted to App Version
	for name, service := range cmp.ServiceGroup {
		if _, err := service.ToVersion(cmp.Name, r.driver.ClusterName()); err != nil {
			http.Error(w, fmt.Sprintf("%s: version convert error: %v", name, err), http.StatusBadRequest)
			return
		}
	}

	// ensure proxy uniq & not occupied
	for _, service := range cmp.ServiceGroup {
		// ensure proxy Listen & Alias uniq
		if err := r.checkProxyDuplication(service.Extra.Proxy); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		// ensure os ports not in using
		if err := r.checkPortListening(service.Extra.Proxy); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
	}

	// get runas
	var runAs = cmp.RunAs()

	// db save
	cmp.ID = utils.RandomString(16)
	cmp.DisplayName = fmt.Sprintf("%s.%s.%s", cmp.Name, runAs, r.driver.ClusterName())
	cmp.OpStatus = types.OpStatusCreating
	cmp.CreatedAt = time.Now()

	if err := r.db.CreateCompose(&cmp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	srvOrders, err := cmp.ServiceGroup.PrioritySort()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		var err error

		// defer mark compose op status
		defer func() {
			if err != nil {
				log.Errorf("launch compose %s error: %v", cmp.Name, err)
				r.memoComposeStatus(cmp.ID, types.OpStatusNoop, err.Error())
			} else {
				log.Printf("launch compose %s succeed", cmp.Name)
				r.memoComposeStatus(cmp.ID, types.OpStatusNoop, "")
			}
		}()

		log.Printf("Preparing to launch compose %s", cmp.Name)

		// create each App by order
		for _, name := range srvOrders {
			var (
				service = cmp.ServiceGroup[name]
				ver, _  = service.ToVersion(cmp.Name, r.driver.ClusterName())
				cluster = r.driver.ClusterName()
				appId   = fmt.Sprintf("%s.%s.%s.%s", ver.Name, cmp.Name, ver.RunAs, cluster)
				count   = int(ver.Instances)
			)

			log.Printf("launching compose app %s with %d tasks", appId, count)

			app := &types.Application{
				ID:        appId,
				Name:      ver.Name,
				RunAs:     ver.RunAs,
				Cluster:   cluster,
				OpStatus:  types.OpStatusCreating,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err = r.db.CreateApp(app); err != nil {
				err = fmt.Errorf("create App %s db App error: %v", appId, err)
				return
			}

			if err = r.db.CreateVersion(appId, ver); err != nil {
				err = fmt.Errorf("create App %s db Version error: %v", appId, err)
				r.memoAppStatus(appId, types.OpStatusNoop, err.Error())
				return
			}

			for i := 0; i < count; i++ {
				var (
					taskName = fmt.Sprintf("%d.%s", i, appId)
					taskId   = fmt.Sprintf("%s.%s", utils.RandomString(12), taskName)
				)

				// db task
				task := &types.Task{
					ID:      taskId,
					Name:    taskName,
					Weight:  100,
					Status:  "pending",
					Healthy: types.TaskHealthyUnset,
					Version: ver.ID,
					Created: time.Now(),
					Updated: time.Now(),
				}

				if ver.IsHealthSet() {
					task.Healthy = types.TaskUnHealthy
				}

				log.Debugf("Create task %s in db", taskId)
				if err = r.db.CreateTask(appId, task); err != nil {
					err = fmt.Errorf("create db task failed: %s", err)
					r.memoAppStatus(appId, types.OpStatusNoop, err.Error())
					return
				}

				// runtime task
				var (
					cfg   = types.NewTaskConfig(ver, i)
					t     = mesos.NewTask(cfg, taskId, taskName)
					tasks = []*mesos.Task{t}
				)

				err = r.driver.LaunchTasks(tasks)
				if err != nil {
					err = fmt.Errorf("launch compose tasks %s error: %v", taskName, err)
					r.memoAppStatus(appId, types.OpStatusNoop, err.Error())
					return
				}
			}

			// max wait for 5 seconds to confirm the preivous app get normal
			if err = r.ensureAppReady(appId, time.Second*5); err != nil {
				r.memoAppStatus(appId, types.OpStatusNoop, err.Error())
				return
			}

			// mark app status
			r.memoAppStatus(appId, types.OpStatusNoop, "")
			log.Printf("compose app %s launch succeed", appId)
		}
	}()

	writeJSON(w, http.StatusCreated, map[string]string{"Id": cmp.ID})
}

func (r *Server) parseYAML(w http.ResponseWriter, req *http.Request) {
	var param struct {
		YAML string `json:"yaml"`
	}

	if err := decode(req.Body, &param); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg, err := utils.YamlServices([]byte(param.YAML), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		vars = utils.YamlVariables([]byte(param.YAML))
		srvs = make([]string, 0, 0)
	)
	for _, srv := range cfg.Services {
		srvs = append(srvs, srv.Name)
	}

	if len(srvs) == 0 {
		http.Error(w, "at least one of docker service definition required", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"services":  srvs,
		"variables": vars,
	})
}

func (r *Server) listComposes(w http.ResponseWriter, req *http.Request) {
	cs, err := r.db.ListComposes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sort.Sort(types.ComposeSorter(cs))
	writeJSON(w, http.StatusOK, cs)
}

func (r *Server) getCompose(w http.ResponseWriter, req *http.Request) {
	composeId := mux.Vars(req)["compose_id"]
	cmp, err := r.db.GetCompose(composeId)
	if err != nil {
		if strings.Contains(err.Error(), "no such compose") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wrapper, err := r.wrapCompose(cmp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, wrapper)
}

func (r *Server) getComposeDependency(w http.ResponseWriter, req *http.Request) {
	composeId := mux.Vars(req)["compose_id"]
	cmp, err := r.db.GetCompose(composeId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dependency, _ := cmp.ServiceGroup.DependMap()
	writeJSON(w, http.StatusOK, dependency)
}

func (r *Server) deleteCompose(w http.ResponseWriter, req *http.Request) {
	composeId := mux.Vars(req)["compose_id"]
	cmp, err := r.db.GetCompose(composeId)
	if err != nil {
		if strings.Contains(err.Error(), "no such compose") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// mark db status
	if err := r.memoComposeStatus(cmp.ID, types.OpStatusDeleting, ""); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		var err error

		// defer mark db status
		defer func() {
			if err != nil {
				err = fmt.Errorf("remove compose %s error: %v", cmp.Name, err)
				log.Errorln(err)
				r.memoComposeStatus(cmp.ID, types.OpStatusNoop, err.Error())
			} else {
				log.Printf("compose %s remove succeed", cmp.Name)
			}
		}()

		log.Printf("Preparing to remove compose %s", cmp.Name)

		// remove each of app
		for name := range cmp.ServiceGroup {
			appId := fmt.Sprintf("%s.%s.%s.%s", name, cmp.Name, cmp.RunAs(), r.driver.ClusterName())

			log.Printf("removing compose app %s ...", appId)

			if _, err = r.db.GetApp(appId); err != nil {
				if r.db.IsErrNotFound(err) {
					log.Printf("removing skip non-exists compose app %s ...", appId)
					continue
				}
				err = fmt.Errorf("get App %s error: %v", appId, err)
				return
			}

			tasks, err := r.db.ListTasks(appId)
			if err != nil {
				err = fmt.Errorf("get App %s tasks error: %v", appId, err)
				return
			}

			versions, err := r.db.ListVersions(appId)
			if err != nil {
				err = fmt.Errorf("get App %s versions error: %v", appId, err)
				return
			}

			if err = r.delApp(appId, tasks, versions); err != nil {
				return
			}

			log.Printf("removed compose app %s", appId)
		}

		// remove db compose
		if err = r.db.DeleteCompose(composeId); err != nil {
			err = fmt.Errorf("remove db compose error: %v", err)
			return
		}
	}()

	writeJSON(w, http.StatusNoContent, "")
}

// short hands to memo update Compose.OpStatus & Compose.ErrMsg
// it's the caller responsibility to process the db error.
func (r *Server) memoComposeStatus(cmpId, op, errmsg string) error {
	cmp, err := r.db.GetCompose(cmpId)
	if err != nil {
		log.Errorf("memoComposeStatus() get db compose %s error: %v", cmpId, err)
		return err
	}

	var (
		prevOp = cmp.OpStatus
	)

	cmp.OpStatus = op
	cmp.ErrMsg = errmsg
	cmp.UpdatedAt = time.Now()

	if err := r.db.UpdateCompose(cmp); err != nil {
		log.Errorf("memoComposeStatus() update app compose status from %s -> %s error: %v",
			prevOp, op, err)
		return err
	}

	return nil
}

// sometimes mesos executor emit consecutive events TASKRUNNING and TASKFAILED if without healthy check,
// which will cause the scheduler treats the task launching as succeed, actually it does NOT.
// we use this to make sure the previous app is ready within compose launching
func (r *Server) ensureAppReady(appId string, maxWait time.Duration) error {
	var (
		app *types.Application
		err error
	)

	for goesby := int64(0); goesby <= int64(maxWait); goesby += int64(time.Second) {
		time.Sleep(time.Second)
		if app, err = r.db.GetApp(appId); err != nil {
			continue
		}

		if app.TasksStatus["TASK_RUNNING"] == app.TaskCount {
			return nil
		}
	}

	if err != nil {
		return err
	}

	return fmt.Errorf("app %s not ready, only %d/%d tasks running",
		appId, app.TasksStatus["TASK_RUNNING"], app.TaskCount)
}

func (r *Server) wrapCompose(cmp *types.Compose) (*types.ComposeWrapper, error) {
	wrapper := &types.ComposeWrapper{Compose: cmp}

	apps, err := r.db.ListApps()
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.Name+"."+cmp.DisplayName == app.ID {
			wrapper.Apps = append(wrapper.Apps, app)
		}
	}
	return wrapper, nil
}
