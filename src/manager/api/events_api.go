package api

import (
	"net/http"
	"strings"

	"github.com/Dataman-Cloud/swan/src/config"
	eventbus "github.com/Dataman-Cloud/swan/src/event"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/manager/scheduler"

	restful "github.com/emicklei/go-restful"
	uuid "github.com/satori/go.uuid"
)

type EventsService struct {
	Scheduler *scheduler.Scheduler
}

func NewAndInstallEventsService(apiServer *apiserver.ApiServer, eng *scheduler.Scheduler) {
	statsService := &EventsService{
		Scheduler: eng,
	}
	apiserver.Install(apiServer, statsService)
}

func (api *EventsService) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		ApiVersion(config.API_PREFIX).
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
	response.Header().Set("Content-Type", "text/event-stream")
	response.Header().Set("Cache-Control", "no-cache")
	response.Write(nil)
	if f, ok := response.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
	if eventbus.Full() {
		http.Error(response, "too many event clients", 500)
		return
	}

	appId := request.QueryParameter("appId")
	catchUp := request.QueryParameter("catchUp")

	listener, _ := eventbus.NewSSEListener(uuid.NewV4().String(), appId, http.ResponseWriter(response))
	eventbus.AddListener(listener)
	defer eventbus.RemoveListener(listener)

	if strings.ToLower(catchUp) == "true" {
		go func() {
			for _, e := range api.Scheduler.HealthyTaskEvents() {
				listener.Write(e)
			}
		}()
	}

	listener.Wait()
}
