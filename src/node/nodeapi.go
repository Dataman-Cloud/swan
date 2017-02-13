package node

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/swancontext"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/utils/httpclient"

	"github.com/Sirupsen/logrus"
	restful "github.com/emicklei/go-restful"
	"golang.org/x/net/context"
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
		Writes([]types.Node{}).
		Filter(swancontext.Instance().ApiServer.Proxy()))

	ws.Route(ws.DELETE("/{node_id}").To(metrics.InstrumentRouteFunc("DELETE", "nodes", api.RemoveNode)).
		Doc("Remove node").
		Operation("removeNode").
		Returns(204, "OK", nil).
		Returns(404, "NotFound", nil).
		Returns(403, "Forbidden remove", nil).
		Param(ws.PathParameter("node_id", "identifier of the node").DataType("string")).
		Filter(swancontext.Instance().ApiServer.Proxy()))

	ws.Route(ws.PATCH("/stop").To(metrics.InstrumentRouteFunc("PATCH", "nodes", api.StopNode)).
		Doc("Stop node").
		Operation("StopNode").
		Returns(200, "OK", nil))

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

	if err := api.node.manager.AddNode(node); err != nil {
		logrus.Errorf("Add node failed, Error: %s", err.Error())
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, api.node.manager.GetNodes())
	return
}

func (api *NodeApi) GetNodes(request *restful.Request, response *restful.Response) {
	response.WriteEntity(api.node.manager.GetNodes())
	return
}

func (api *NodeApi) RemoveNode(request *restful.Request, response *restful.Response) {
	targetNodeID := request.PathParameter("node_id")
	if targetNodeID == api.node.ID {
		logrus.Errorf("Remove the leader node is forbidden")
		response.WriteErrorString(http.StatusForbidden, "can not remove leader")
		return
	}

	targetNode, err := api.node.manager.GetNode(targetNodeID)
	if err != nil {
		logrus.Errorf("Remove node: %s failed, Error: %s", targetNodeID, err.Error())
		response.WriteError(http.StatusNotFound, err)
		return
	}

	targetNodeAddr := "http://" + targetNode.AdvertiseAddr + config.API_PREFIX + "/nodes/stop"
	_, err = httpclient.NewDefaultClient().PATCH(context.TODO(), targetNodeAddr, nil, nil, nil)
	if err != nil {
		logrus.Errorf("Remove node: %s failed, Error: %s", targetNodeID, err.Error())
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	if err := api.node.manager.RemoveNode(targetNode); err != nil {
		logrus.Errorf("Remove node: %s failed, Error: %s", targetNodeID, err.Error())
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusNoContent, nil)
	return
}

func (api *NodeApi) StopNode(request *restful.Request, response *restful.Response) {
	response.WriteHeaderAndEntity(http.StatusOK, nil)
	api.node.Stop()
	return
}
