package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/apiserver"
	"github.com/Dataman-Cloud/swan/src/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"
	//"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/manager/framework/store" // TODO(xychu): better not touch store in API layer

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"golang.org/x/net/context"
)

type QuotaService struct {
	Store     store.Store
	apiserver.ApiRegister
}

func NewAndInstallQuotaService(apiServer *apiserver.ApiServer, store store.Store) *QuotaService {
	quotaService := &QuotaService{
		Store: store,
	}
	apiserver.Install(apiServer, quotaService)
	return quotaService
}

// NOTE(xychu): Every service need to registed to ApiServer need to impl
//              a `Register` interface so that it can be added to ApiServer.Start
func (api *QuotaService) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		ApiVersion(config.API_PREFIX).
		Path(config.API_PREFIX + "/quotas").
		Doc("Quota management").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(metrics.InstrumentRouteFunc("GET", "Quotas", api.ListQuotas)).
		// docs
		Doc("List Quotas").
		Operation("listQuotas").
		Returns(200, "OK", []types.ResourceQuota{}))
	ws.Route(ws.POST("/").To(metrics.InstrumentRouteFunc("POST", "Quota", api.CreateQuota)).
		// docs
		Doc("Create Quota").
		Operation("createQuota").
		Returns(201, "OK", types.ResourceQuota{}).
		Returns(400, "BadRequest", nil).
		Reads(types.ResourceQuota{}).
		Writes(types.ResourceQuota{}))
	ws.Route(ws.GET("/{quota_group}").To(metrics.InstrumentRouteFunc("GET", "Quota", api.GetQuota)).
		// docs
		Doc("Get an Quota").
		Operation("getQuota").
		Param(ws.PathParameter("quota_group", "identifier of the quota group").DataType("string")).
		Returns(200, "OK", types.ResourceQuota{}).
		Returns(404, "NotFound", nil).
		Writes(types.ResourceQuota{}))
	ws.Route(ws.DELETE("/{quota_group}").To(metrics.InstrumentRouteFunc("DELETE", "Quota", api.DeleteQuota)).
		// docs
		Doc("Delete Quota").
		Operation("deleteQuota").
		Returns(204, "OK", nil).
		Returns(404, "NotFound", nil).
		Param(ws.PathParameter("quota_group", "identifier of the quota group").DataType("string")))
	ws.Route(ws.PUT("/{quota_group}").To(metrics.InstrumentRouteFunc("PUT", "Quota", api.UpdateQuota)).
		// docs
		Doc("Update Quota").
		Operation("updateQuota").
		Returns(200, "OK", types.ResourceQuota{}).
		Returns(404, "NotFound", nil).
		Reads(types.ResourceQuota{}).
		Writes(types.ResourceQuota{}).
		Param(ws.PathParameter("quota_group", "identifier of the quota group").DataType("string")))

	container.Add(ws)
}

func (api *QuotaService) CreateQuota(request *restful.Request, response *restful.Response) {
	var quota types.ResourceQuota

	err := request.ReadEntity(&quota)
	if err != nil {
		logrus.Errorf("Create quota error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	err = api.Store.CreateQuota(context.TODO(), &quota, nil)
	if err != nil {
		logrus.Errorf("Create quota error: %s", err.Error())
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, quota)
}

func (api *QuotaService) ListQuotas(request *restful.Request, response *restful.Response) {
	quotas, err := api.Store.ListQuotas()
	if err != nil {
		logrus.Errorf("List quotas error: %s", err.Error())
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	response.WriteEntity(quotas)
}

func (api *QuotaService) GetQuota(request *restful.Request, response *restful.Response) {
	quota, err := api.Store.GetQuota(request.PathParameter("quota_group"))
	if err != nil {
		logrus.Errorf("Get quota error: %s", err.Error())
		response.WriteError(http.StatusNotFound, err)
		return
	}
	response.WriteEntity(quota)
}

func (api *QuotaService) DeleteQuota(request *restful.Request, response *restful.Response) {
	err := api.Store.DeleteQuota(context.TODO(), request.PathParameter("quota_group"), nil)
	if err != nil {
		logrus.Errorf("Delete quota error: %s", err.Error())
		response.WriteError(http.StatusNotFound, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (api *QuotaService) UpdateQuota(request *restful.Request, response *restful.Response) {
	var quota types.ResourceQuota

	err := request.ReadEntity(&quota)
	if err != nil {
		logrus.Errorf("Update quota error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	quotaGroup := request.PathParameter("quota_group")
	err = api.Store.UpdateQuota(context.TODO(), &quota, nil)
	if err != nil {
		logrus.Errorf("Update quota[%s] error: %s", quotaGroup, err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	quotaRet, err := api.Store.GetQuota(quotaGroup)
	if err != nil {
		logrus.Errorf("Inspect quota[%s] error: %s", quotaGroup, err.Error())
		response.WriteError(http.StatusNotFound, err)
		return
	}
	response.WriteEntity(quotaRet)

}
