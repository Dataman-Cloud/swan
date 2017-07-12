package api

import (
	"net/http"
)

func (r *Server) ping(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, "pong")
}
