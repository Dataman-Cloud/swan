package manager

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/utils/httpclient"

	"github.com/Sirupsen/logrus"
	restful "github.com/emicklei/go-restful"
	"golang.org/x/net/context"
)

type ManagerApi struct {
	manager *Manager
}

func (api *ManagerApi) Register(container *restful.Container) {
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
		Filter(api.manager.apiServer.Proxy()))

	ws.Route(ws.DELETE("/{node_id}").To(metrics.InstrumentRouteFunc("DELETE", "nodes", api.RemoveNode)).
		Doc("Remove node").
		Operation("removeNode").
		Returns(204, "OK", nil).
		Returns(404, "NotFound", nil).
		Returns(403, "Forbidden remove", nil).
		Param(ws.PathParameter("node_id", "identifier of the node").DataType("string")).
		Filter(api.manager.apiServer.Proxy()))

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

func (api *ManagerApi) AddNode(request *restful.Request, response *restful.Response) {
	var node types.Node

	if err := request.ReadEntity(&node); err != nil {
		logrus.Errorf("Add node failed, Error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	if err := api.manager.AddNode(node); err != nil {
		logrus.Errorf("Add node failed, Error: %s", err.Error())
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, api.manager.GetNodes())
	return
}

func (api *ManagerApi) GetNodes(request *restful.Request, response *restful.Response) {
	response.WriteEntity(api.manager.GetNodes())
	return
}

func (api *ManagerApi) RemoveNode(request *restful.Request, response *restful.Response) {
	targetNodeID := request.PathParameter("node_id")
	if targetNodeID == api.manager.NodeInfo.ID {
		logrus.Errorf("Remove the leader node is forbidden")
		response.WriteErrorString(http.StatusForbidden, "can not remove leader")
		return
	}

	targetNode, err := api.manager.GetNode(targetNodeID)
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

	if err := api.manager.RemoveNode(targetNode); err != nil {
		logrus.Errorf("Remove node: %s failed, Error: %s", targetNodeID, err.Error())
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusNoContent, "success")
	return
}

func (api *ManagerApi) StopNode(request *restful.Request, response *restful.Response) {
	response.WriteHeaderAndEntity(http.StatusOK, "success")
	api.manager.Stop()
	return
}
