package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"
	log "github.com/Sirupsen/logrus"
	uuid "github.com/satori/go.uuid"
)

func (r *Server) newCompose(w http.ResponseWriter, req *http.Request) {
	if err := checkForJSON(req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var cps types.Compose
	if err := decode(req.Body, &cps); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// check conflict
	if cps, _ := r.db.GetCompose(cps.Name); cps != nil {
		http.Error(w, fmt.Sprintf("compose %s already exists", cps.Name), http.StatusConflict)
		return
	}

	if cps.RequireConvert() {
		sgroup, err := cps.ToServiceGroup()

		if err != nil {
			http.Error(w, fmt.Sprintf("yaml convert: %v", err), http.StatusBadRequest)
			return
		}

		cps.ServiceGroup = sgroup
	}

	// verify
	if err := cps.Valid(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get runas
	var runAs string
	for _, ext := range cps.YAMLExtra {
		if ext != nil {
			runAs = ext.RunAs
			break
		}
	}

	// db save
	cps.ID = uuid.NewV4().String()
	cps.DisplayName = fmt.Sprintf("%s.%s.%s", cps.Name, runAs, r.driver.ClusterName())
	cps.Status = "creating"
	cps.CreatedAt = time.Now()
	cps.UpdatedAt = time.Now()

	if err := r.db.CreateCompose(&cps); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	srvOrders, err := cps.ServiceGroup.PrioritySort()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		for _, srv := range srvOrders {
			ver, err := cps.ServiceGroup[srv].ToVersion(cps.Name, r.driver.ClusterName())
			if err != nil {
				log.Errorf("convert compose to version got error: %v", err)
				return
			}

			var (
				id  = fmt.Sprintf("%s.%s.%s.%s", ver.Name, cps.Name, ver.RunAs, r.driver.ClusterName())
				vid = fmt.Sprintf("%d", time.Now().UTC().UnixNano())
			)

			app := &types.Application{
				ID:        id,
				Name:      ver.Name,
				Alias:     ver.Proxy.Alias,
				RunAs:     ver.RunAs,
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

			ver.ID = vid

			if err := r.db.CreateVersion(id, ver); err != nil {
				http.Error(w, fmt.Sprintf("create app version failed: %v", err), http.StatusInternalServerError)
				return
			}

			count := int(ver.Instances)

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
					Status:  "creating",
					Healthy: types.TaskHealthyUnset,
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

				tasks := []*mesos.Task{t}

				results, err := r.driver.LaunchTasks(tasks)
				if err != nil {
					log.Errorf("launch task %s got error: %v", id, err)

					task.Status = "Failed"
					task.ErrMsg = err.Error()

					if err = r.db.UpdateTask(app.ID, task); err != nil {
						log.Errorf("update task %s got error: %v", id, err)
					}

					break
				}

				for taskId, err := range results {
					if err != nil {
						log.Errorf("launch task %s got error: %v", taskId, err)
						break
					}
				}

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
		http.Error(w, "at least one of docker service defination required", http.StatusBadRequest)
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
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cps, err := r.db.GetCompose(req.Form.Get("compose_id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, cps)
}

func (r *Server) deleteCompose(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := r.db.DeleteCompose(req.Form.Get("compose_id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, "OK")
}
