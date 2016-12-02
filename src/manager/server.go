package manager

import (
	"encoding/json"
	"net/http"

	"github.com/Dataman-Cloud/swan/src/manager/apiserver/router"
)

type Router struct {
	routes  []*router.Route
	manager *Manager
}

func NewRouter(manager *Manager) *Router {
	r := &Router{
		manager: manager,
	}

	r.initRoutes()
	return r
}

func (r *Router) initRoutes() {
	r.routes = []*router.Route{
		router.NewRoute("GET", "/v1/manager/state", r.State),
	}
}

type ManagerState struct {
	FrameworkId string
	RaftId      int
	LeaderId    uint64
	Cluster     string
}

func (r *Router) Routes() []*router.Route {
	return r.routes
}

func (r *Router) State(w http.ResponseWriter, req *http.Request) error {
	managerState := &ManagerState{
		RaftId:  r.manager.config.Raft.RaftId,
		Cluster: r.manager.config.Raft.Cluster,
	}

	frameworkId, err := r.manager.swanContext.Store.FetchFrameworkID()
	if err != nil {
		return err
	}
	managerState.FrameworkId = frameworkId

	leader, err := r.manager.raftNode.Leader()
	if err != nil {
		return err
	}

	managerState.LeaderId = leader
	return json.NewEncoder(w).Encode(managerState)
}
