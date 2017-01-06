package manager

import (
	"encoding/json"
	"net/http"

	"github.com/Dataman-Cloud/swan/src/swancontext"
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
		RaftId:  swancontext.Instance().Config.Raft.RaftId,
		Cluster: swancontext.Instance().Config.Raft.Cluster,
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
