package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"

	"github.com/emicklei/go-restful"
	"github.com/satori/go.uuid"
)

type EventsService struct {
	Scheduler *scheduler.Scheduler
	apiserver.ApiRegister
}

func NewAndInstallEventsService(apiServer *apiserver.ApiServer, eng *scheduler.Scheduler) *EventsService {
	statsService := &EventsService{
		Scheduler: eng,
	}
	apiserver.Install(apiServer, statsService)
	return statsService
}

func (api *EventsService) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		ApiVersion(API_PREFIX).
		Path("/events").
		Doc("events API").
		Produces(restful.MIME_JSON).
		Produces("*/*")

	ws.Route(ws.GET("/").To(metrics.InstrumentRouteFunc("GET", "Events", api.Events)).
		// docs
		Doc("Get Events").
		Operation("getEvents").
		Param(ws.QueryParameter("appId", "appId, e.g. appId=nginx0051").DataType("string")).
		Returns(200, "OK", ""))

	container.Add(ws)
}

func (api *EventsService) Events(request *restful.Request, response *restful.Response) {
	appId := request.QueryParameter("appId")
	subscriber, doneChan := event.NewSSESubscriber(uuid.NewV4().String(), appId, http.ResponseWriter(response))
	subscriber.Subscribe(swancontext.Instance().EventBus)
	<-doneChan
}
