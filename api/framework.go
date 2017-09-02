package api

import (
	"net/http"
)

func (r *Server) getFrameworkInfo(w http.ResponseWriter, req *http.Request) {
	info := r.driver.FrameworkInfo()
	writeJSON(w, http.StatusOK, info)
}
