package api

import (
	"fmt"
	"net/http"

	"github.com/Dataman-Cloud/swan/types"

	"github.com/gorilla/mux"
)

func (s *Server) listMesosAgents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.driver.ListAgents())
}

func (s *Server) updateMesosAgent(w http.ResponseWriter, r *http.Request) {
	agentId := mux.Vars(r)["agent_id"]

	label := new(types.MesosLabel)
	if err := decode(r.Body, label); err != nil {
		http.Error(w, fmt.Sprintf("decode label got error: %v", err), http.StatusBadRequest)
		return
	}

	agent, err := s.db.GetMesosAgent(agentId)
	if err != nil {
		if s.db.IsErrNotFound(err) {
			agent, err := s.newMesosAgent(agentId, label)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			writeJSON(w, http.StatusOK, agent)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.labelExists(agent.Attrs, label.Key) {
		http.Error(w, fmt.Sprintf("decode label got error: %v", err), http.StatusConflict)
		return
	}

	agent.Attrs[label.Key] = label.Value

	if err := s.db.UpdateMesosAgent(agent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, agent.Attrs)
}

func (s *Server) newMesosAgent(agentId string, label *types.MesosLabel) (*types.MesosAgent, error) {
	agent := &types.MesosAgent{
		ID:    agentId,
		Attrs: make(map[string]string),
	}

	agent.Attrs[label.Key] = label.Value

	if err := s.db.CreateMesosAgent(agent); err != nil {
		return nil, err
	}

	return agent, nil
}

func (s *Server) labelExists(labels map[string]string, key string) bool {
	for k := range labels {
		if k == key {
			return true
		}
	}

	return false
}
