package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
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

	// get runas
	var runAs = cmp.RunAs()

	// db save
	cmp.ID = uuid.NewV4().String()
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
			cmp.OpStatus = types.OpStatusNoop
			cmp.UpdatedAt = time.Now()
			if err != nil {
				cmp.ErrMsg = err.Error()
				log.Errorf("launch compose %s error: %v", cmp.Name, err)
			}
			if err := r.db.UpdateCompose(&cmp); err != nil {
				log.Errorf("update db compose %s error: %v", cmp.Name, err)
			}
		}()

		// create each App by order
		for _, name := range srvOrders {
			var (
				service = cmp.ServiceGroup[name]
				ver, _  = service.ToVersion(cmp.Name, r.driver.ClusterName())
				cluster = r.driver.ClusterName()
				id      = fmt.Sprintf("%s.%s.%s.%s", ver.Name, cmp.Name, ver.RunAs, cluster)
			)

			app := &types.Application{
				ID:        id,
				Name:      ver.Name,
				RunAs:     ver.RunAs,
				Cluster:   cluster,
				OpStatus:  types.OpStatusCreating,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err = r.db.CreateApp(app); err != nil {
				err = fmt.Errorf("create App %s db App error: %v", name, err)
				return
			}

			if err = r.db.CreateVersion(id, ver); err != nil {
				err = fmt.Errorf("create App %s db Version error: %v", name, err)
				return
			}

			count := int(ver.Instances)
			log.Debugf("Preparing to launch %d App %s tasks", count, name)

			for i := 0; i < count; i++ {
				var (
					name = fmt.Sprintf("%d.%s", i, app.ID)
					id   = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
				)

				cfg := types.NewTaskConfig(ver)

				if cfg.Network != "host" && cfg.Network != "bridge" {
					cfg.Parameters = append(cfg.Parameters, &types.Parameter{
						Key:   "ip",
						Value: ver.IPs[i],
					})

					cfg.IP = ver.IPs[i]
				}

				task := &types.Task{
					ID:      id,
					Name:    name,
					Weight:  100,
					Status:  "pending",
					Healthy: types.TaskHealthyUnset,
					Version: ver.ID,
					Created: time.Now(),
					Updated: time.Now(),
				}

				healthSet := ver.HealthCheck != nil && !ver.HealthCheck.IsEmpty()
				if healthSet {
					task.Healthy = types.TaskUnHealthy
				}

				log.Debugf("Create task %s in db", task.ID)
				if err = r.db.CreateTask(app.ID, task); err != nil {
					err = fmt.Errorf("create db task failed: %s", err)
					return
				}

				t := mesos.NewTask(
					cfg,
					id,
					name,
				)

				var (
					tasks   = []*mesos.Task{t}
					results map[string]error
				)

				results, err = r.driver.LaunchTasks(tasks)
				if err != nil {
					err = fmt.Errorf("launch compose App %s tasks error: %v", name, err)
					return
				}

				if n := len(results); n > 0 {
					err = fmt.Errorf("launch %d compose tasks of App %s error: %v", n, name, results)
					// TODO memo errmsg within errored App.ErrMsg
					return
				}
			}

			// mark app status
			app.OpStatus = types.OpStatusNoop
			if err := r.db.UpdateApp(app); err != nil {
				log.Errorf("update app op-status got error: %v", err)
			}

		}
	}()

	writeJSON(w, http.StatusAccepted, "accepted")
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

	writeJSON(w, http.StatusOK, cs)
}

func (r *Server) getCompose(w http.ResponseWriter, req *http.Request) {
	composeId := mux.Vars(req)["compose_id"]
	cmp, err := r.db.GetCompose(composeId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, cmp)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// mark db status
	cmp.OpStatus = types.OpStatusDeleting
	if err := r.db.UpdateCompose(cmp); err != nil {
		log.Errorf("update db compose %s error: %v", cmp.ID, err)
	}

	go func() {
		var err error

		// defer mark db status
		defer func() {
			if err != nil {
				log.Errorf("remove compose %s error: %v", cmp.Name, err)
				cmp.OpStatus = types.OpStatusNoop
				cmp.UpdatedAt = time.Now()
				cmp.ErrMsg = err.Error()
				if err := r.db.UpdateCompose(cmp); err != nil {
					log.Errorf("update db compose %s error: %v", cmp.ID, err)
				}
			}
		}()

		// remove each of app
		for name := range cmp.ServiceGroup {
			appId := fmt.Sprintf("%s.%s.%s.%s", name, cmp.Name, cmp.RunAs(), r.driver.ClusterName())

			if _, err := r.db.GetApp(appId); err != nil {
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
		}

		// remove db compose
		if err = r.db.DeleteCompose(composeId); err != nil {
			err = fmt.Errorf("remove db compose error: %v", err)
			return
		}
	}()

	writeJSON(w, http.StatusNoContent, "")
}
