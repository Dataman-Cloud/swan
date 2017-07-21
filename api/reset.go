package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func (r *Server) resetStatus(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["app_id"]
	app, err := r.db.GetApp(id)
	if err != nil {
		if strings.Contains(err.Error(), "node does not exist") {
			http.Error(w, fmt.Sprintf("app %s not exists", id), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.OpStatus = types.OpStatusNoop
	if err := r.db.UpdateApp(app); err != nil {
		log.Errorf("reset app's op-status to noop got error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"opStatus": "noop"})
}
