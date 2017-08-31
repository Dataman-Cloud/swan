package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"

	"github.com/gorilla/mux"
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
	vclusterId := mux.Vars(r)["vcluster_id"]

	if err := s.db.DeleteVCluster(vclusterId); err != nil {
		if !s.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusNoContent, "")
}

func (s *Server) addNode(w http.ResponseWriter, r *http.Request) {
	vclusterId := mux.Vars(r)["vcluster_id"]

	_, err := s.db.GetVCluster(vclusterId)
	if err != nil {
		if s.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	body := new(types.CreateNodeBody)
	if err := decode(r.Body, body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	node := &types.Node{
		ID:    body.ID,
		IP:    body.IP,
		Attrs: make(map[string]string),
	}

	if err := s.db.CreateNode(vclusterId, node); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, node.ID)
}

func (s *Server) updateNode(w http.ResponseWriter, r *http.Request) {
	var (
		vars       = mux.Vars(r)
		vclusterId = vars["vcluster_id"]
		nodeId     = vars["node_id"]
	)

	node, err := s.db.GetNode(vclusterId, nodeId)
	if err != nil {
		if s.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	label := new(types.MesosLabel)
	if err := decode(r.Body, label); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if utils.LabelExists(node.Attrs, label.Key) {
		http.Error(w, fmt.Sprintf("label %s=%s already exists", label.Key, label.Value), http.StatusConflict)
		return
	}

	node.Attrs[label.Key] = label.Value

	if err := s.db.UpdateNode(vclusterId, node); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, node)
}

// func (s *Server) labelExists(labels map[string]string, key string) bool {
// 	for k := range labels {
// 		if k == key {
// 			return true
// 		}
// 	}
//
// 	return false
// }
