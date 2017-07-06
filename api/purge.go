package api

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/Dataman-Cloud/swan/types"
	log "github.com/Sirupsen/logrus"
)

func (r *Router) purge(w http.ResponseWriter, req *http.Request) {
	apps, err := r.db.ListApps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenBucket := make(chan struct{}, 50) // TODO(nmg): delete step, make it configurable

	go func() {
		var all sync.WaitGroup

		for _, app := range apps {
			var (
				hasError = false
				wg       sync.WaitGroup
			)

			wg.Add(len(app.Tasks))
			for _, task := range app.Tasks {
				tokenBucket <- struct{}{}

				go func(task *types.Task, appId string) {
					all.Add(1)

					defer func() {
						wg.Done()
						all.Done()
						<-tokenBucket
					}()

					if err := r.driver.KillTask(task.ID, task.AgentId); err != nil {
						log.Errorf("Kill task %s got error: %v", task.ID, err)

						hasError = true

						task.OpStatus = fmt.Sprintf("kill task error: %v", err)
						if err = r.db.UpdateTask(appId, task); err != nil {
							log.Errorf("update task %s got error: %v", task.Name, err)
						}

						return
					}

					if err := r.db.DeleteTask(task.ID); err != nil {
						log.Errorf("Delete task %s got error: %v", task.ID, err)

						hasError = true

						task.OpStatus = fmt.Sprintf("delete task error: %v", err)
						if err = r.db.UpdateTask(appId, task); err != nil {
							log.Errorf("update task %s got error: %v", task.Name, err)
						}

						return
					}

				}(task, app.ID)
			}

			wg.Wait()

			if hasError {
				log.Errorf("Delete some tasks of app %s got error.", app.ID)
				return
			}

			if err := r.db.DeleteApp(app.ID); err != nil {
				log.Error("Delete app %s got error: %v", app.ID, err)
				return
			}

		}

		all.Wait()

		close(tokenBucket)
	}()

	writeJSON(w, http.StatusNoContent, "")
}
