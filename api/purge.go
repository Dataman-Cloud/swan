package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/types"
	log "github.com/Sirupsen/logrus"
)

func (r *Server) purge(w http.ResponseWriter, req *http.Request) {
	apps, err := r.db.ListApps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cmps, err := r.db.ListComposes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		// remove apps
		for _, app := range apps {
			var appId = app.ID

			tasks, err := r.db.ListTasks(appId)
			if err != nil {
				log.Errorf("Purge() list app %s tasks error: %v", appId, err)
				continue
			}

			versions, err := r.db.ListVersions(appId)
			if err != nil && !r.db.IsErrNotFound(err) {
				log.Errorf("Purge() list app %s versions error: %v", appId, err)
				continue
			}
			// mark app op status
			if err := r.memoAppStatus(appId, types.OpStatusDeleting, ""); err != nil {
				log.Errorf("Purge() update app opstatus to deleting got error: %v", err)
				continue
			}

			if err = r.delApp(appId, tasks, versions); err != nil {
				log.Errorf("Purge() delte app %s error: %v", appId, err)
				continue
			}
		}

		// remove composes
		for _, cmp := range cmps {
			if err := r.db.DeleteCompose(cmp.ID); err != nil {
				log.Errorf("Purege() remove db compose %s error: %v", cmp.ID, err)
				continue
			}
		}
	}()

	writeJSON(w, http.StatusNoContent, "")
}
