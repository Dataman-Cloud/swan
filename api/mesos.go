package api

import (
	"fmt"
	"net/http"

	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func (s *Server) listMesosAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := s.db.ListMesosAgents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, agents)
}

func (s *Server) getAgentLabels(w http.ResponseWriter, r *http.Request) {
	agentIp := mux.Vars(r)["agent_ip"]

	agent, err := s.db.GetMesosAgent(agentIp)
	if err != nil {
		if s.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, agent)
}

func (s *Server) createAgentLabel(w http.ResponseWriter, r *http.Request) {
	agentIp := mux.Vars(r)["agent_ip"]

	label := new(types.MesosLabel)
	if err := decode(r.Body, label); err != nil {
		http.Error(w, fmt.Sprintf("decode label got error: %v", err), http.StatusBadRequest)
		return
	}

	if err := label.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	agent, err := s.db.GetMesosAgent(agentIp)
	if err != nil {
		if s.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.labelExists(agent.Attrs, label.Key) {
		http.Error(w, fmt.Sprintf("duplicated label : %s", label.Key), http.StatusConflict)
		return
	}

	agent.Attrs[label.Key] = label.Value

	if err := s.db.UpdateMesosAgent(agent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, agent)
}

func (s *Server) updateAgentLabel(w http.ResponseWriter, r *http.Request) {
	agentIp := mux.Vars(r)["agent_ip"]

	label := new(types.MesosLabel)
	if err := decode(r.Body, label); err != nil {
		http.Error(w, fmt.Sprintf("decode label got error: %v", err), http.StatusBadRequest)
		return
	}

	if !s.canOperated(label) {
		http.Error(w, fmt.Sprintf("label %s=%s is in used", label.Key, label.Value), http.StatusLocked)
		return
	}

	agent, err := s.db.GetMesosAgent(agentIp)
	if err != nil {
		if s.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !s.labelExists(agent.Attrs, label.Key) {
		http.Error(w, fmt.Sprintf("no such label: %s", label.Key), http.StatusNotFound)
		return
	}

	agent.Attrs[label.Key] = label.Value

	if err := s.db.UpdateMesosAgent(agent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, agent)
}

func (s *Server) deleteAgentLabel(w http.ResponseWriter, r *http.Request) {
	agentIp := mux.Vars(r)["agent_ip"]

	label := new(types.MesosLabel)
	if err := decode(r.Body, label); err != nil {
		http.Error(w, fmt.Sprintf("decode label got error: %v", err), http.StatusBadRequest)
		return
	}

	if !s.canOperated(label) {
		http.Error(w, fmt.Sprintf("label %s=%s is in used", label.Key, label.Value), http.StatusLocked)
		return
	}

	agent, err := s.db.GetMesosAgent(agentIp)
	if err != nil {
		if s.db.IsErrNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !s.labelExists(agent.Attrs, label.Key) {
		writeJSON(w, http.StatusNoContent, "")
		return
	}

	delete(agent.Attrs, label.Key)

	if err := s.db.UpdateMesosAgent(agent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusNoContent, "")
}

func (s *Server) newMesosAgent(agentIp string, label *types.MesosLabel) (*types.MesosAgent, error) {
	agent := &types.MesosAgent{
		ID:    agentIp,
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

func (s *Server) canOperated(label *types.MesosLabel) bool {
	apps, err := s.db.ListApps()
	if err != nil {
		log.Errorf("canOperated list apps error: %v", err)
		return false
	}

	for _, app := range apps {
		ver, err := s.db.GetVersion(app.ID, app.Version[0])
		if err != nil {
			log.Errorf("canOperated get version failed: %v", err)
			return false
		}

		constraints := ver.Constraints
		log.Println("===", constraints)
		if len(constraints) == 0 {
			continue
		}

		for _, constraint := range constraints {
			if constraint.Attribute == label.Key {
				return false
			}
		}
	}

	return true
}
