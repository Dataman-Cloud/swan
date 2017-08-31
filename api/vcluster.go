package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"
)

func (s *Server) listVClusters(w http.ResponseWriter, r *http.Request) {
	vcs, err := s.db.ListVClusters()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, vcs)
}

func (s *Server) createVCluster(w http.ResponseWriter, r *http.Request) {
	if err := checkForJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body := new(types.CreateVClusterBody)
	if err := decode(r.Body, body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if s.db.VClusterExists(body.Name) {
		http.Error(w, fmt.Sprintf("vcluster %s has already exists.", body.Name), http.StatusConflict)
		return
	}

	vcluster := &types.VCluster{
		ID:      utils.RandomString(32),
		Name:    body.Name,
		Created: time.Now(),
		Updated: time.Now(),
	}

	if err := s.db.CreateVCluster(vcluster); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"ID": vcluster.ID})
}

func (s *Server) deleteVCluster(w http.ResponseWriter, r *http.Request) {
	vclusterId := mux.Vars(r)["cluster_id"]

	if err := s.db.DeleteVCluster(vclusterId); err != nil {
		if !r.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusNoContent, "")
}
