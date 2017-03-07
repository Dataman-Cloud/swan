package agent

import (
	"net/http"

	"github.com/Dataman-Cloud/swan-janitor/src"
	"github.com/Dataman-Cloud/swan-resolver/nameserver"
	"github.com/Dataman-Cloud/swan/src/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/event"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Sirupsen/logrus"

	restful "github.com/emicklei/go-restful"
	"github.com/mitchellh/mapstructure"
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

	ws.Route(ws.POST("/init").To(metrics.InstrumentRouteFunc("POST", "Agent Init", api.Init)).
		Doc("Init Agent").
		Operation("InitAgent").
		Returns(201, "OK", nil).
		Returns(400, "BadRequest", nil).
		Reads([]nameserver.RecordGeneratorChangeEvent{}).
		Writes(nil))

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

func (api *AgentApi) Init(request *restful.Request, response *restful.Response) {
	var events []*event.Event
	if err := request.ReadEntity(&events); err != nil {
		logrus.Errorf("init agent failed. Error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	go api.dispenseEvents(events)

	response.WriteHeaderAndEntity(http.StatusCreated, nil)
	return
}

func (api *AgentApi) dispenseEvents(events []*event.Event) {
	for _, taskEvent := range events {
		// taskEvent.Payload is type of interface{} can not auto unmarshal
		var eventInfo types.TaskInfoEvent
		mapstructure.Decode(taskEvent.Payload, &eventInfo)
		taskEvent.Payload = &eventInfo

		resolverEvent, err := event.BuildResolverEvent(taskEvent)
		if err == nil {
			logrus.Infof("agent got resolver event %+v", resolverEvent)
			api.agent.resolver.RecordGeneratorChangeChan() <- resolverEvent
		} else {
			logrus.Errorf("Build resolver event got error: %s", err.Error())
		}

		janitorEvent, err := event.BuildJanitorEvent(taskEvent)
		if err == nil {
			logrus.Infof("agent got resolver event %+v", janitorEvent)
			api.agent.janitorServer.SwanEventChan() <- janitorEvent
		} else {
			logrus.Errorf("Build janitor event got error: %s", err.Error())
		}
	}
}

func (api *AgentApi) ResolverEventHandler(request *restful.Request, response *restful.Response) {
	var resolverEvent nameserver.RecordGeneratorChangeEvent
	if err := request.ReadEntity(&resolverEvent); err != nil {
		logrus.Errorf("handle resolver event failed. Error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	logrus.Infof("agent got resolver event %+v", resolverEvent)
	api.agent.resolver.RecordGeneratorChangeChan() <- &resolverEvent

	response.WriteHeaderAndEntity(http.StatusCreated, nil)
	return
}

func (api *AgentApi) JanitorEventHandler(request *restful.Request, response *restful.Response) {
	var janitorEvent janitor.TargetChangeEvent
	if err := request.ReadEntity(&janitorEvent); err != nil {
		logrus.Errorf("handle janitor event failed. Error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	logrus.Infof("agent got janitor event %+v", janitorEvent)
	api.agent.janitorServer.SwanEventChan() <- &janitorEvent

	response.WriteHeaderAndEntity(http.StatusCreated, nil)
	return

}
