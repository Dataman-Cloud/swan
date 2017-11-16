package api

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// Kvm App
//
//

func (r *Server) listKvmApps(w http.ResponseWriter, req *http.Request) {
	apps, err := r.db.ListKvmApps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, apps)
}

func (r *Server) createKvmApp(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	// obtian query parameters about identifications
	var (
		name    = req.Form.Get("name")
		runAs   = req.Form.Get("runas")
		cluster = req.Form.Get("cluster")
		desc    = req.Form.Get("desc")
	)
	if cluster == "" {
		cluster = r.driver.ClusterName()
	}

	// obtian configs
	var config types.KvmConfig
	if err := decode(req.Body, &config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var id = fmt.Sprintf("%s.%s.%s", name, runAs, cluster)
	// check conflict
	if app, _ := r.db.GetKvmApp(id); app != nil {
		http.Error(w, fmt.Sprintf("kvm app %s already exists", id), http.StatusConflict)
		return
	}

	// prepare object
	kvmApp := &types.KvmApp{
		ID:        id,
		Name:      name,
		RunAs:     runAs,
		Cluster:   cluster,
		Desc:      desc,
		Config:    &config,
		OpStatus:  types.OpStatusCreating,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// verify
	if err := kvmApp.Valid(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// db create kvm app
	if err := r.db.CreateKvmApp(kvmApp); err != nil {
		if strings.Contains(err.Error(), "app already exists") {
			http.Error(w, fmt.Sprintf("kvm app %s has already exists", id), http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("create db kvm app error: %v", err), http.StatusInternalServerError)
		return
	}

	// create & run kvm tasks async
	go func(appId string) {
		var err error

		// defer to mark op status
		defer func() {
			if err != nil {
				log.Errorf("launch kvm app %s error: %v", appId, err)
				r.memoKvmAppStatus(appId, types.OpStatusNoop, fmt.Sprintf("launch kvm app error: %v", err))
			} else {
				log.Printf("launch kvm app %s succeed", appId)
				r.memoKvmAppStatus(appId, types.OpStatusNoop, "")
			}
		}()

		// prepare for all runtime tasks & db tasks
		log.Printf("Preparing to launch Kvm App %s with %d tasks", appId, config.Count)

		tasks := []*mesos.Task{}
		for i := 0; i < config.Count; i++ {
			var (
				name = fmt.Sprintf("%d.%s", i, appId)
				id   = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
			)

			task := &types.KvmTask{
				ID:        id,
				Name:      name,
				Status:    "pending",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err = r.db.CreateKvmTask(appId, task); err != nil {
				err = fmt.Errorf("create db kvm task failed: %v", err)
				return
			}

			// runtime kvm tasks
			tasks = append(tasks, mesos.NewKvmTask(id, name, &config))
		}

		// wait scheduler launching all tasks
		err = r.driver.LaunchTasks(tasks)
		if err != nil {
			err = fmt.Errorf("launch tasks got error: %v", err)
			return
		}

	}(kvmApp.ID)

	writeJSON(w, http.StatusCreated, map[string]string{"id": kvmApp.ID})
}

func (r *Server) getKvmApp(w http.ResponseWriter, req *http.Request) {
	var (
		appId = mux.Vars(req)["app_id"]
	)

	app, err := r.db.GetKvmApp(appId)
	if err != nil {
		if r.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, app)
}

func (r *Server) deleteKvmApp(w http.ResponseWriter, req *http.Request) {
	var (
		appId = mux.Vars(req)["app_id"]
	)

	// get app
	app, err := r.db.GetKvmApp(appId)
	if err != nil {
		if r.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get app tasks
	tasks, err := r.db.ListKvmTasks(appId)
	if err != nil {
		http.Error(w, fmt.Sprintf("list kvm tasks got error for delete kvm app. %v", err), http.StatusInternalServerError)
		return
	}

	log.Debugf("kvm app %s has %d tasks", appId, len(tasks))

	// mark app op status
	if err := r.memoKvmAppStatus(appId, types.OpStatusDeleting, ""); err != nil {
		http.Error(w, fmt.Sprintf("update kvm app opstatus to deleting got error: %v", err), http.StatusInternalServerError)
		return
	}

	// preform kill & remove async
	go func() {
		var err error

		// defer to mark op status
		defer func() {
			if err != nil {
				log.Errorf("delete kvm app %s error: %v", appId, err)
				r.memoKvmAppStatus(appId, types.OpStatusNoop, fmt.Sprintf("delete kvm app error: %v", err))
			} else {
				log.Printf("delete kvm app %s succeed", appId)
			}
		}()

		log.Printf("Preparing to delete Kvm App %s with %d tasks", appId, len(tasks))

		// kill runtime tasks & remove db tasks
		var (
			count   = len(tasks)
			succeed = int64(0)
			wg      sync.WaitGroup
		)
		for _, task := range tasks {
			wg.Add(1)
			go func(task *types.KvmTask) {
				defer wg.Done()

				if err := r.delKvmTask(appId, app.Config.KillPolicy, task); err != nil {
					return
				}

				atomic.AddInt64(&succeed, 1)
			}(task)
		}
		wg.Wait()

		if int(succeed) != count {
			err = fmt.Errorf("%d tasks kill / removed failed", count-int(succeed))
			return
		}

		// remove db app
		err = r.db.DeleteKvmApp(appId)
	}()

	writeJSON(w, http.StatusNoContent, "")
}

func (r *Server) stopKvmApp(w http.ResponseWriter, req *http.Request) {
	r.controlKvmApp(w, req, "Stop")
}

func (r *Server) startKvmApp(w http.ResponseWriter, req *http.Request) {
	r.controlKvmApp(w, req, "Start")
}

func (r *Server) suspendKvmApp(w http.ResponseWriter, req *http.Request) {
	r.controlKvmApp(w, req, "Suspend")
}

func (r *Server) resumeKvmApp(w http.ResponseWriter, req *http.Request) {
	r.controlKvmApp(w, req, "Resume")
}

func (r *Server) controlKvmApp(w http.ResponseWriter, req *http.Request, ops string) {
	var (
		appId = mux.Vars(req)["app_id"]
	)

	// get app tasks
	tasks, err := r.db.ListKvmTasks(appId)
	if err != nil {
		http.Error(w, fmt.Sprintf("list kvm tasks got error. %v", err), http.StatusInternalServerError)
		return
	}

	log.Debugf("kvm app %s has %d tasks", appId, len(tasks))

	// op type
	var (
		opStatus string
		opFunc   func(string, string, string) error
	)
	switch ops {
	case "Start":
		opStatus = types.OpStatusStarting
		opFunc = r.driver.StartKvmTask
	case "Stop":
		opStatus = types.OpStatusStopping
		opFunc = r.driver.StopKvmTask
	case "Suspend":
		opStatus = types.OpStatusSuspending
		opFunc = r.driver.SuspendKvmTask
	case "Resume":
		opStatus = types.OpStatusResuming
		opFunc = r.driver.ResumeKvmTask
	default:
		http.Error(w, "unsupported kvm app operation", http.StatusBadRequest)
		return
	}

	// mark app op status
	if err := r.memoKvmAppStatus(appId, opStatus, ""); err != nil {
		http.Error(w, fmt.Sprintf("update kvm app opstatus to %s got error: %v", opStatus, err), http.StatusInternalServerError)
		return
	}

	// preform start async
	go func() {
		var err error

		// defer to mark op status
		defer func() {
			if err != nil {
				log.Errorf("%s kvm app %s error: %v", ops, appId, err)
				r.memoKvmAppStatus(appId, types.OpStatusNoop, fmt.Sprintf("%s kvm app error: %v", ops, err))
			} else {
				log.Printf("%s kvm app %s succeed", ops, appId)
			}
		}()

		log.Printf("Preparing to %s Kvm App %s with %d tasks", ops, appId, len(tasks))

		// start runtime tasks
		var (
			count   = len(tasks)
			succeed = int64(0)
			wg      sync.WaitGroup
		)
		for _, task := range tasks {
			wg.Add(1)
			go func(task *types.KvmTask) {
				defer wg.Done()

				if err = opFunc(task.ID, task.AgentId, task.ExecutorId); err != nil {
					log.Errorf("%s kvm task %s got error: %v", ops, task.ID, err)
					r.memoKvmTaskStatus(appId, task.ID, "", "", ops+" kvm task error: "+err.Error())
					return
				}

				atomic.AddInt64(&succeed, 1)

			}(task)
		}
		wg.Wait()

		if int(succeed) != count {
			err = fmt.Errorf("%d tasks %s failed", count-int(succeed), ops)
			return
		}
	}()

	writeJSON(w, http.StatusNoContent, "")
}

// Kvm Task
//
//
func (r *Server) listKvmTasks(w http.ResponseWriter, req *http.Request) {
	var (
		appId = mux.Vars(req)["app_id"]
	)

	tasks, err := r.db.ListKvmTasks(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (r *Server) getKvmTask(w http.ResponseWriter, req *http.Request) {
	var (
		vars   = mux.Vars(req)
		appId  = vars["app_id"]
		taskId = vars["task_id"]
	)

	task, err := r.db.GetKvmTask(appId, taskId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (r *Server) deleteKvmTask(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) startKvmTask(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) stopKvmTask(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) suspendKvmTask(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) resumeKvmTask(w http.ResponseWriter, req *http.Request) {
}

// Related Utils
//
//

// delKvmTask actually kill runtime kvm task & remove db objects
func (r *Server) delKvmTask(appId string, policy *types.KillPolicy, task *types.KvmTask) error {
	var gracePeriod int64
	if policy != nil {
		gracePeriod = policy.Duration
	}

	if err := r.driver.KillTask(task.ID, task.AgentId, gracePeriod); err != nil {
		log.Errorf("Kill kvm task %s got error: %v", task.ID, err)
		r.memoKvmTaskStatus(appId, task.ID, "", "", "kill kvm task error: "+err.Error())
		return err
	}

	if err := r.db.DeleteKvmTask(task.ID); err != nil {
		log.Errorf("Delete kvm task %s got error: %v", task.ID, err)
		r.memoKvmTaskStatus(appId, task.ID, "", "", "delete kvm task error: "+err.Error())
		return err
	}

	return nil
}

// short hands to memo update App.OpStatus & App.ErrMsg
// it's the caller responsibility to process the db error.
func (r *Server) memoKvmAppStatus(appId, op, errmsg string) error {
	app, err := r.db.GetKvmApp(appId)
	if err != nil {
		log.Errorf("memoKvmAppStatus() get db kvm app %s error: %v", appId, err)
		return err
	}

	var (
		prevOp = app.OpStatus
	)

	app.OpStatus = op
	app.ErrMsg = errmsg
	app.UpdatedAt = time.Now()

	if err := r.db.UpdateKvmApp(app); err != nil {
		log.Errorf("memoKvmAppStatus() update kvm app db status from %s -> %s error: %v", prevOp, op, err)
		return err
	}

	return nil
}

// similar as above, but for kvm task
func (r *Server) memoKvmTaskStatus(appId, taskId, op, status, errmsg string) error {
	task, err := r.db.GetKvmTask(appId, taskId)
	if err != nil {
		log.Errorf("memoKvmTaskStatus() get db kvm task %s error: %v", taskId, err)
		return err
	}

	task.OpStatus = op
	task.Status = status // mesos task status
	task.ErrMsg = errmsg // mesos task errmsg
	task.UpdatedAt = time.Now()

	if err := r.db.UpdateKvmTask(appId, task); err != nil {
		log.Errorf("memoKvmTaskStatus() update db kvm task %s error: %v", taskId, err)
		return err
	}

	return nil
}
