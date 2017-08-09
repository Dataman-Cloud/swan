package api

import (
	"fmt"
	"net/http"

	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func (r *Server) resetStatus(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["app_id"]
	app, err := r.db.GetApp(id)
	if err != nil {
		if r.db.IsErrNotFound(err) {
			http.Error(w, fmt.Sprintf("app %s not exists", id), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var (
		current = app.OpStatus
		desired = types.OpStatusNoop
	)

	app.OpStatus = desired
	if err := r.db.UpdateApp(app); err != nil {
		log.Errorf("reset app's op-status to noop got error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"previous": current,
		"current":  desired,
	})
}
