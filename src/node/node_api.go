package node

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
	restful "github.com/emicklei/go-restful"
)

type NodeApi struct {
	node *Node
}

func (api *NodeApi) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		ApiVersion(config.API_PREFIX).
		Path(config.API_PREFIX + "/nodes").
		Doc("manager server api").
		Consumes(restful.MIME_JSON).
		Produces("*/*")

	ws.Route(ws.POST("/").To(metrics.InstrumentRouteFunc("POST", "nodes", api.AddNode)).
		Doc("Add node").
		Operation("AddNode").
		Returns(201, "OK", []types.Node{}).
		Returns(400, "BadRequest", nil).
		Reads(types.Node{}).
		Writes([]types.Node{}))

	ws.Route(ws.GET("/").To(metrics.InstrumentRouteFunc("GET", "nodes", api.GetNodes)).
		Doc("Get Nodes").
		Operation("GetNodes").
		Returns(200, "OK", []types.Node{}))

	container.Add(ws)
}

func (api *NodeApi) AddNode(request *restful.Request, response *restful.Response) {
	var node types.Node

	if err := request.ReadEntity(&node); err != nil {
		logrus.Errorf("Add node failed, Error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	if node.IsAgent() {
		if err := api.node.manager.AddAgent(node); err != nil {
			logrus.Errorf("Add agent failed, Error: %s", err.Error())
			response.WriteError(http.StatusInternalServerError, err)
			return
		}
	}

	response.WriteHeaderAndEntity(http.StatusCreated, api.node.manager.GetNodes())
	return
}

func (api *NodeApi) GetNodes(request *restful.Request, response *restful.Response) {
	response.WriteEntity(api.node.manager.GetNodes())
	return
}
