package api

import (
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

	for _, app := range apps {
		go func(app *types.Application) {
			var (
				chErrors = make(chan error, 0)
				wg       sync.WaitGroup
			)
			wg.Add(len(app.Tasks))
			for _, task := range app.Tasks {
				go func(task *types.Task) {
					defer wg.Done()

					if err := r.driver.KillTask(task.ID, task.AgentId); err != nil {
						chErrors <- err
						return
					}

					if err := r.db.DeleteTask(task.ID); err != nil {
						chErrors <- err
						return
					}

				}(task)
			}

			wg.Wait()

			if len(chErrors) == 0 {
				if err := r.db.DeleteApp(app.ID); err != nil {
					// TODO(nmg): should show failed reason.
					log.Error("Delete app %s got error: %v", app.ID, err)
					return
				}
			}

		}(app)
	}

	writeJSON(w, http.StatusNoContent, "")
}
