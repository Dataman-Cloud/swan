package api

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/Dataman-Cloud/swan/types"
	log "github.com/Sirupsen/logrus"
)

func (r *Server) purge(w http.ResponseWriter, req *http.Request) {
	apps, err := r.db.ListApps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		for _, app := range apps {
			go func(appId string) {
				tasks, err := r.db.ListTasks(appId)
				if err != nil {
					log.Errorf("list app tasks got error for purge: %v", err)
					return
				}

				var (
					count         = len(tasks)
					succeed int64 = 0
					wg      sync.WaitGroup
				)

				for _, task := range tasks {
					wg.Add(1)
					go func(task *types.Task) {
						defer wg.Done()

						if err := r.driver.KillTask(task.ID, task.AgentId, true); err != nil {
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

				if int(succeed) == count {
					versions, err := r.db.ListVersions(appId)
					if err != nil {
						log.Errorf("list versions error for delete app. %v", err)
						return
					}

					for _, version := range versions {
						if err := r.db.DeleteVersion(appId, version.ID); err != nil {
							log.Errorf("Delete version %s for app %s got error: %v", version.ID, appId, err)
							return
						}
					}

					if err := r.db.DeleteApp(appId); err != nil {
						log.Errorf("Delete app %s got error: %v", appId, err)
					}
				}
			}(app.ID)
		}
	}()

	writeJSON(w, http.StatusNoContent, "")
}
