package api

import (
	"net/http"
)

func (r *Router) ping(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, "pong")
}
