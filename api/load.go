package api

import "net/http"

func (r *Router) load(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, r.driver.Load())
}
