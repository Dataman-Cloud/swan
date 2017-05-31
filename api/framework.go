package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/types"
)

func (r *Router) getFrameworkInfo(w http.ResponseWriter, req *http.Request) {
	info := new(types.FrameworkInfo)

	writeJSON(w, http.StatusOK, info)
}
