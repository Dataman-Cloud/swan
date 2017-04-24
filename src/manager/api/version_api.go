package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/version"

	restful "github.com/emicklei/go-restful"
)

type VersionService struct {
}

func NewAndInstallVersionService(apiServer *apiserver.ApiServer) {
	apiserver.Install(apiServer, new(VersionService))
}

func (api *VersionService) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		ApiVersion(config.API_PREFIX).
		Path("/version").
		Doc("version API").
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(metrics.InstrumentRouteFunc("GET", "Version", api.Version)).
		Doc("Version").
		Operation("version").
		Returns(200, "OK", version.Version{}))

	container.Add(ws)
}

func (api *VersionService) Version(request *restful.Request, response *restful.Response) {
	response.WriteHeaderAndEntity(http.StatusOK, version.GetVersion())
}
