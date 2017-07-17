package api

import "net/http"

func (r *Server) load(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, r.driver.Load())
}
