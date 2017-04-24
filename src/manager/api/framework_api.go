package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/manager/connector"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/emicklei/go-restful"
)

type FrameworkService struct {
}

func NewAndInstallFrameworkService(apiServer *apiserver.ApiServer) {
	apiserver.Install(apiServer, new(FrameworkService))
}

func (fs *FrameworkService) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		ApiVersion(config.API_PREFIX).
		Path(config.API_PREFIX + "/framework").
		Doc("framework info").
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/info").To(metrics.InstrumentRouteFunc("GET", "Info", fs.Info)).
		Doc("Info").
		Operation("info").
		Returns(200, "OK", types.FrameworkInfo{}))

	container.Add(ws)
}

func (fs *FrameworkService) Info(req *restful.Request, resp *restful.Response) {
	info := new(types.FrameworkInfo)
	if c := connector.Instance(); c != nil {
		info.ID = c.FrameworkInfo.GetId().GetValue()
	} else {
		info.ID = ""
	}

	resp.WriteHeaderAndEntity(http.StatusOK, info)
}
