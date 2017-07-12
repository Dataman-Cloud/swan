package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/types"
)

func (r *Server) getLeader(w http.ResponseWriter, req *http.Request) {
	leader := &types.Leader{
		Leader: r.GetLeader(),
	}

	writeJSON(w, http.StatusOK, leader)
}
