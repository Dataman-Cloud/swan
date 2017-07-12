package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/version"
)

func (r *Server) version(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, version.GetVersion())
}
