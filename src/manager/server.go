package manager

import (
	"encoding/json"
	"net/http"
)

type Router struct {
	manager *Manager
}

func NewRouter(manager *Manager) *Router {
	r := &Router{
		manager: manager,
	}

	return r
}

type ManagerState struct {
	FrameworkId string
	RaftId      int
	LeaderId    uint64
	Cluster     string
}

func (r *Router) State(w http.ResponseWriter, req *http.Request) error {
	managerState := &ManagerState{
		RaftId:  r.manager.config.Raft.RaftId,
		Cluster: r.manager.config.Raft.Cluster,
	}

	//frameworkId, err := r.manager.swanContext.Store.FetchFrameworkID()
	//if err != nil {
	//	return err
	//}
	//managerState.FrameworkId = frameworkId

	leader, err := r.manager.raftNode.Leader()
	if err != nil {
		return err
	}

	managerState.LeaderId = leader
	return json.NewEncoder(w).Encode(managerState)
}
