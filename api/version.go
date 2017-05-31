package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/version"
)

func (r *Router) version(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, version.GetVersion())
}
