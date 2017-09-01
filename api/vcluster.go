package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"

	"github.com/gorilla/mux"
)

func (s *Server) getVCluster(w http.ResponseWriter, r *http.Request) {
	v, err := s.db.GetVCluster(mux.Vars(r)["vcluster_name"])
	if err != nil {
		if s.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, v)
}

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

	writeJSON(w, http.StatusCreated, vcluster)
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
	name := mux.Vars(r)["vcluster_name"]

	_, err := s.db.GetVCluster(name)
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
		ID:    utils.RandomString(32),
		IP:    body.IP,
		Attrs: make(map[string]string),
	}

	node.Attrs["cluster"] = name

	if err := s.db.CreateNode(name, node); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, node)
}

func (s *Server) updateNode(w http.ResponseWriter, r *http.Request) {
	var (
		vars   = mux.Vars(r)
		name   = vars["vcluster_name"]
		nodeIp = vars["node_ip"]
	)

	node, err := s.db.GetNode(name, nodeIp)
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

	node.Attrs[label.Key] = label.Value

	if err := s.db.UpdateNode(name, node); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, node)
}

func (s *Server) listNodes(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["vcluster_name"]

	nodes, err := s.db.ListNodes(name)
	if err != nil {
		if s.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	writeJSON(w, http.StatusOK, nodes)
}

func (s *Server) getNode(w http.ResponseWriter, r *http.Request) {
	var (
		vars   = mux.Vars(r)
		name   = vars["vcluster_name"]
		nodeIp = vars["node_ip"]
	)

	node, err := s.db.GetNode(name, nodeIp)
	if err != nil {
		if s.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, node)
}

func (s *Server) delNode(w http.ResponseWriter, r *http.Request) {
	var (
		vars   = mux.Vars(r)
		name   = vars["vcluster_name"]
		nodeIp = vars["node_ip"]
	)

	if err := s.db.DeleteNode(name, nodeIp); err != nil {
		if !s.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusNoContent, "")
}
