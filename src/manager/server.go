package manager

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
	restful "github.com/emicklei/go-restful"
)

type ManagerApi struct {
	manager *Manager
}

func (api *ManagerApi) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		ApiVersion(config.API_PREFIX).
		Path(config.API_PREFIX + "/manager/agents").
		Doc("manager server api").
		Consumes(restful.MIME_JSON).
		Produces("*/*")

	ws.Route(ws.POST("/").To(metrics.InstrumentRouteFunc("POST", "Agents", api.AddAgent)).
		Doc("Add Agent").
		Operation("AddAgent").
		Returns(201, "OK", types.Agent{}).
		Returns(400, "BadRequest", nil).
		Reads(types.Agent{}).
		Writes(types.Agent{}))

	ws.Route(ws.GET("/").To(metrics.InstrumentRouteFunc("GET", "Agents", api.GetAgents)).
		Doc("Get Agents").
		Operation("GetAgents").
		Returns(200, "OK", []types.Agent{}))

	container.Add(ws)
}

func (api *ManagerApi) AddAgent(request *restful.Request, response *restful.Response) {
	var agent types.Agent

	if err := request.ReadEntity(&agent); err != nil {
		logrus.Errorf("Add agent failed, Error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	if err := api.manager.AddAgent(agent); err != nil {
		logrus.Errorf("Add agent failed, Error: %s", err.Error())
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, agent)
	return
}

func (api *ManagerApi) GetAgents(request *restful.Request, response *restful.Response) {
	var agents []types.Agent
	for _, agent := range api.manager.GetAgents() {
		agents = append(agents, agent)
	}
	response.WriteEntity(agents)
	return
}

type ManagerState struct {
	FrameworkId string
	RaftId      int
	LeaderId    uint64
	Cluster     string
}

//func (r *Router) State(w http.ResponseWriter, req *http.Request) error {
//	managerState := &ManagerState{
//		RaftId:  swancontext.Instance().Config.Raft.RaftId,
//		Cluster: swancontext.Instance().Config.Raft.Cluster,
//	}
//
//	//frameworkId, err := r.manager.swanContext.Store.FetchFrameworkID()
//	//if err != nil {
//	//	return err
//	//}
//	//managerState.FrameworkId = frameworkId
//
//	leader, err := r.manager.raftNode.Leader()
//	if err != nil {
//		return err
//	}
//
//	managerState.LeaderId = leader
//	return json.NewEncoder(w).Encode(managerState)
//}
