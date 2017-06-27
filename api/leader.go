package api

import (
	"fmt"
	"net/http"

	"github.com/Dataman-Cloud/swan/types"
)

func (r *Router) getLeader(w http.ResponseWriter, req *http.Request) {

	lead, err := r.db.GetLeader()
	if err != nil {
		http.Error(w, fmt.Sprintf("Find leader got error: %v", err), http.StatusInternalServerError)
		return
	}

	leader := &types.Leader{
		Leader: lead,
	}

	writeJSON(w, http.StatusOK, leader)
}
