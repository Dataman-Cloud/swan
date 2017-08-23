package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func (r *Server) runComposeNG(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var (
		name    = req.Form.Get("name")
		cluster = req.Form.Get("cluster")
		runAs   = req.Form.Get("runas")
		desc    = req.Form.Get("desc")
		envs    = req.Form.Get("envs")   // k1=v1,k2=v2,k3=v3
		labels  = req.Form.Get("labels") // k1=v1,k2=v2,k3=v3
	)
	if cluster == "" {
		cluster = r.driver.ClusterName()
	}

	// check conflict
	if cmp, _ := r.db.GetComposeNG(name); cmp != nil {
		http.Error(w, fmt.Sprintf("compose %s already exists", name), http.StatusConflict)
		return
	}

	// obtain extra labels
	extLabels := make(map[string]string)
	for _, pair := range strings.Split(labels, ",") {
		if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
			extLabels[kv[0]] = kv[1]
		}
	}

	// new compose app
	var cmpApp = &types.ComposeApp{
		ID:          utils.RandomString(16),
		Name:        name,
		RunAs:       runAs,
		Cluster:     cluster,
		DisplayName: fmt.Sprintf("%s.%s.%s", name, runAs, cluster),
		Desc:        desc,
		OpStatus:    types.OpStatusCreating,
		CreatedAt:   time.Now(),
		Labels:      extLabels,
	}

	// obtain yaml text
	yaml, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cmpApp.YAMLRaw = string(yaml)

	// obtain yaml envs
	envMap := make(map[string]string)
	for _, pair := range strings.Split(envs, ",") {
		if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
			envMap[kv[0]] = kv[1]
		}
	}
	cmpApp.YAMLEnv = envMap

	// parse yaml text to types.ComposeV3
	cmp, err := types.ParseComposeV3(yaml, envMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cmpApp.ComposeV3 = cmp

	// verify
	if err := cmpApp.Valid(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ensure every service could be converted to app Version
	convertedVers, err := cmpApp.ParseComposeToVersions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ensure proxy uniq & not occupied
	for _, svr := range cmp.Services {
		// ensure proxy Listen & Alias uniq
		if err := r.checkProxyDuplication(svr.Proxy); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		// ensure os ports not in using
		if err := r.checkPortListening(svr.Proxy); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
	}

	// services priority sort
	svrOrders, err := cmp.PrioritySort()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// db save
	if err := r.db.CreateComposeNG(cmpApp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// launch compose app
	go func() {
		var err error

		// defer mark compose op status
		defer func() {
			if err != nil {
				log.Errorf("launch compose %s error: %v", cmpApp.Name, err)
				r.memoComposeStatusNG(cmpApp.ID, types.OpStatusNoop, err.Error())
			} else {
				log.Printf("launch compose %s succeed", cmpApp.Name)
				r.memoComposeStatusNG(cmpApp.ID, types.OpStatusNoop, "")
			}
		}()

		log.Printf("Preparing to launch compose %s: %v", cmpApp.Name, svrOrders)

		// create each App by order
		for _, name := range svrOrders {
			var (
				service   = cmp.Services[name]
				ver       = convertedVers[name]
				waitDelay = service.WaitDelay
				appId     = fmt.Sprintf("%s.%s.%s.%s", ver.Name, cmpApp.Name, ver.RunAs, cluster)
				count     = int(ver.Instances)
			)

			if ver == nil {
				err = fmt.Errorf("converted app version not found for %s", name)
				return
			}

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

			// wait delay before next service
			if waitDelay > 0 {
				time.Sleep(time.Second * time.Duration(waitDelay))
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

	writeJSON(w, http.StatusCreated, map[string]string{"Id": cmpApp.ID})
}

func (r *Server) parseYAMLNG(w http.ResponseWriter, req *http.Request) {
	yaml, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cmp, err := types.ParseComposeV3(yaml, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"services":  cmp.GetServices(),
		"variables": cmp.GetVariables(),
	})
}

func (r *Server) listComposesNG(w http.ResponseWriter, req *http.Request) {
	cs, err := r.db.ListComposesNG()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.ParseForm()
	cs = r.filterComposeNG(cs, req.Form)

	sort.Sort(types.ComposeAppSorter(cs))
	writeJSON(w, http.StatusOK, cs)
}

func (r *Server) getComposeNG(w http.ResponseWriter, req *http.Request) {
	var (
		composeId = mux.Vars(req)["compose_id"]
	)

	cmpApp, err := r.db.GetComposeNG(composeId)
	if err != nil {
		if r.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wrapper, err := r.wrapComposeNG(cmpApp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, wrapper)
}

func (r *Server) getComposeDependencyNG(w http.ResponseWriter, req *http.Request) {
	var (
		composeId = mux.Vars(req)["compose_id"]
	)

	cmpApp, err := r.db.GetComposeNG(composeId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dependency, _ := cmpApp.ComposeV3.DependMap()
	writeJSON(w, http.StatusOK, dependency)
}

func (r *Server) parseComposeToVersions(w http.ResponseWriter, req *http.Request) {
	var (
		composeId = mux.Vars(req)["compose_id"]
	)

	cmpApp, err := r.db.GetComposeNG(composeId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	versions, _ := cmpApp.ParseComposeToVersions()
	writeJSON(w, http.StatusOK, versions)
}

func (r *Server) deleteComposeNG(w http.ResponseWriter, req *http.Request) {
	var (
		composeId = mux.Vars(req)["compose_id"]
	)

	cmpApp, err := r.db.GetComposeNG(composeId)
	if err != nil {
		if r.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// mark db status
	if err := r.memoComposeStatusNG(cmpApp.ID, types.OpStatusDeleting, ""); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		var err error

		// defer mark db status
		defer func() {
			if err != nil {
				err = fmt.Errorf("remove compose %s error: %v", cmpApp.Name, err)
				log.Errorln(err)
				r.memoComposeStatusNG(cmpApp.ID, types.OpStatusNoop, err.Error())
			} else {
				log.Printf("compose %s remove succeed", cmpApp.Name)
			}
		}()

		log.Printf("Preparing to remove compose %s", cmpApp.Name)

		var (
			cmpName = cmpApp.Name
			runAs   = cmpApp.RunAs
			cluster = cmpApp.Cluster
		)

		// remove each of app
		for name := range cmpApp.ComposeV3.Services {
			appId := fmt.Sprintf("%s.%s.%s.%s", name, cmpName, runAs, cluster)

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
		if err = r.db.DeleteComposeNG(composeId); err != nil {
			err = fmt.Errorf("remove db compose error: %v", err)
			return
		}
	}()

	writeJSON(w, http.StatusNoContent, "")
}

// short hands to memo update Compose.OpStatus & Compose.ErrMsg
// it's the caller responsibility to process the db error.
func (r *Server) memoComposeStatusNG(cmpId, op, errmsg string) error {
	cmp, err := r.db.GetComposeNG(cmpId)
	if err != nil {
		log.Errorf("memoComposeStatusNG() get db compose %s error: %v", cmpId, err)
		return err
	}

	var (
		prevOp = cmp.OpStatus
	)

	cmp.OpStatus = op
	cmp.ErrMsg = errmsg
	cmp.UpdatedAt = time.Now()

	if err := r.db.UpdateComposeNG(cmp); err != nil {
		log.Errorf("memoComposeStatusNG() update app compose status from %s -> %s error: %v",
			prevOp, op, err)
		return err
	}

	return nil
}

func (r *Server) wrapComposeNG(cmpApp *types.ComposeApp) (*types.ComposeAppWrapper, error) {
	wrapper := &types.ComposeAppWrapper{ComposeApp: cmpApp}

	apps, err := r.db.ListApps()
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.Name+"."+cmpApp.DisplayName == app.ID {
			wrapper.Apps = append(wrapper.Apps, app)
		}
	}
	return wrapper, nil
}

// filter compose apps according by labels
func (r *Server) filterComposeNG(cmpApps []*types.ComposeApp, filter url.Values) []*types.ComposeApp {
	var idx int

	for _, cmpApp := range cmpApps {
		labels := cmpApp.Labels
		if len(labels) == 0 {
			continue
		}

		var n int
		for k, v := range filter {
			for key, val := range labels {
				if key == k && val == v[0] {
					n++
					break
				}
			}
		}

		// if matched all filters
		if n == len(filter) {
			cmpApps[idx] = cmpApp
			idx++
		}
	}

	return cmpApps[0:idx]
}
