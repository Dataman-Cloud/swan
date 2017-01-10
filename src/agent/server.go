package agent

import (
	"net/http"

	"github.com/Dataman-Cloud/swan-janitor/src/upstream"
	"github.com/Dataman-Cloud/swan-resolver/nameserver"
	"github.com/Dataman-Cloud/swan/src/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/event"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Sirupsen/logrus"

	restful "github.com/emicklei/go-restful"
)

type AgentApi struct {
	agent *Agent
}

func (api *AgentApi) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		ApiVersion(config.API_PREFIX).
		Path(config.API_PREFIX + "/agent").
		Doc("agent server api").
		Consumes(restful.MIME_JSON).
		Produces("*/*")

	ws.Route(ws.POST("/init").To(metrics.InstrumentRouteFunc("POST", "Agent Init", api.InitAgent)).
		Doc("Init Agent").
		Operation("InitAgent").
		Returns(201, "OK", types.Agent{}).
		Returns(400, "BadRequest", nil).
		Reads([]types.App{}).
		Writes(types.Agent{}))

	ws.Route(ws.POST("/resolver/event").To(metrics.InstrumentRouteFunc("POST", "Resolver Event", api.ResolverEventHandler)).
		Doc("Resolver Event Handler").
		Operation("Handle Resolver Event").
		Returns(201, "OK", nil).
		Returns(400, "BadRequest", nil).
		Reads(event.Event{}).
		Writes(nil))

	ws.Route(ws.POST("/janitor/event").To(metrics.InstrumentRouteFunc("POST", "Janitor Event", api.JanitorEventHandler)).
		Doc("Janitor Event Handler").
		Operation("Handler Janitor Event").
		Returns(201, "OK", nil).
		Returns(400, "BadRequest", nil).
		Reads(event.Event{}).
		Writes(nil))

	container.Add(ws)
}

func (api *AgentApi) InitAgent(request *restful.Request, response *restful.Response) {
}

func (api *AgentApi) ResolverEventHandler(request *restful.Request, response *restful.Response) {
	var resolverEvent nameserver.RecordGeneratorChangeEvent
	if err := request.ReadEntity(&resolverEvent); err != nil {
		logrus.Errorf("handle resolver event failed. Error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	api.agent.resolver.RecordGeneratorChangeChan() <- &resolverEvent

	response.WriteHeaderAndEntity(http.StatusCreated, nil)
	return
}

func (api *AgentApi) JanitorEventHandler(request *restful.Request, response *restful.Response) {
	var janitorEvent upstream.TargetChangeEvent
	if err := request.ReadEntity(&janitorEvent); err != nil {
		logrus.Errorf("handle janitor event failed. Error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	api.agent.janitorServer.SwanEventChan() <- &janitorEvent

	response.WriteHeaderAndEntity(http.StatusCreated, nil)
	return

}
