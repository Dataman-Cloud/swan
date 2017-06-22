package api

import (
	"net/http"
	"sync"
)

func (r *Router) purge(w http.ResponseWriter, req *http.Request) {
	apps, err := r.db.ListApps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var wg sync.WaitGroup
	for _, app := range apps {
		wg.Add(1)
		go func() {
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

			if err := r.db.DeleteApp(app.ID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			wg.Done()
		}()

	}

	wg.Wait()

	writeJSON(w, http.StatusOK, "OK")
}
