package api

import (
	"net/http"
)

func (r *Router) leader(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, r.db.GetLeader())
}
