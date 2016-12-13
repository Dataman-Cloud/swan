package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	swanapiserver "github.com/Dataman-Cloud/swan/src/manager/new_apiserver"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

const (
	API_PREFIX = "v_beta"
)

type Api struct {
	config    config.Scheduler
	Scheduler *scheduler.Scheduler
}

func NewApi(eng *scheduler.Scheduler, config config.Scheduler) *Api {
	return &Api{
		Scheduler: eng,
		config:    config,
	}
}

// TODO(xychu): after we port all the apis to restful.WebService,
//              we could remove Api and NewApi above.
type AppService struct {
	Scheduler *scheduler.Scheduler
	swanapiserver.ApiRegister
}

func NewAndInstallAppService(apiServer *swanapiserver.ApiServer, eng *scheduler.Scheduler) *AppService {
	appService := &AppService{
		Scheduler: eng,
	}
	swanapiserver.Install(apiServer, appService)
	return appService
}

// NOTE(xychu): Every service need to registed to ApiServer need to impl
//              a `Register` interface so that it can be added to ApiServer.Start
func (api *AppService) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		ApiVersion(API_PREFIX).
		Path("/apps").
		Doc("App management").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(api.ListApp).
		// docs
		Doc("List Apps").
		Operation("listApps").
		Returns(200, "OK", []types.Application{}))
	ws.Route(ws.POST("/").To(api.CreateApp).
		// docs
		Doc("Create App").
		Operation("createApp").
		Returns(201, "OK", types.Application{}).
		Returns(400, "BadRequest", nil).
		Reads(types.Version{}).
		Writes(types.Application{}))
	ws.Route(ws.GET("/{app_id}").To(api.GetApp).
		// docs
		Doc("Get an App").
		Operation("getApp").
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")).
		Returns(200, "OK", types.Application{}).
		Returns(404, "NotFound", nil).
		Writes(types.Application{}))
	ws.Route(ws.DELETE("/{app_id}").To(api.DeleteApp).
		// docs
		Doc("Delete App").
		Operation("deleteApp").
		Returns(204, "OK", nil).
		Returns(404, "NotFound", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PATCH("/{app_id}/scale-up").To(api.ScaleUp).
		// docs
		Doc("Scale Up App").
		Operation("scaleUp").
		Returns(404, "NotFound", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PATCH("/{app_id}/scale-down").To(api.ScaleDown).
		// docs
		Doc("Scale Down App").
		Operation("scaleDown").
		Returns(404, "NotFound", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PUT("/{app_id}").To(api.UpdateApp).
		// docs
		Doc("Update App").
		Operation("updateApp").
		Returns(404, "NotFound", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))

	container.Add(ws)
}

func (api *AppService) CreateApp(request *restful.Request, response *restful.Response) {
	var version types.Version

	err := request.ReadEntity(&version)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
	}

	err = CheckVersion(&version)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
	}

	app, err := api.Scheduler.CreateApp(&version)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	} else {
		response.WriteHeaderAndEntity(http.StatusCreated, app)
	}
}

func (api *AppService) ListApp(request *restful.Request, response *restful.Response) {
	appsRet := make([]*App, 0)
	for _, app := range api.Scheduler.ListApps() {
		version := app.CurrentVersion
		appsRet = append(appsRet, &App{
			ID:               version.AppId,
			Name:             version.AppId,
			Instances:        int(version.Instances),
			RunningInstances: app.RunningInstances(),
			RunAs:            version.RunAs,
			ClusterId:        app.MesosConnector.ClusterId,
			Created:          app.Created,
			Updated:          app.Updated,
			Mode:             string(app.Mode),
			State:            app.State,
		})
	}

	response.WriteEntity(appsRet)
}

func (api *AppService) GetApp(request *restful.Request, response *restful.Response) {
	app, err := api.Scheduler.InspectApp(request.PathParameter("app_id"))
	if err != nil {
		response.WriteError(http.StatusNotFound, err)
	} else {
		version := app.CurrentVersion
		appRet := &App{
			ID:               version.AppId,
			Name:             version.AppId,
			Instances:        int(version.Instances),
			RunningInstances: app.RunningInstances(),
			RunAs:            version.RunAs,
			ClusterId:        app.MesosConnector.ClusterId,
			Created:          app.Created,
			Updated:          app.Updated,
			Mode:             string(app.Mode),
			State:            app.State,
		}

		appRet.Versions = make([]string, 0)
		for _, v := range app.Versions {
			appRet.Versions = append(appRet.Versions, v.ID)
		}

		appRet.Tasks = FilterTasksFromApp(app)

		response.WriteEntity(appRet)
	}
}

func (api *AppService) DeleteApp(request *restful.Request, response *restful.Response) {
	err := api.Scheduler.DeleteApp(request.PathParameter("app_id"))
	if err != nil {
		response.WriteError(http.StatusNotFound, err)
	} else {
		response.WriteHeader(http.StatusNoContent)
	}
}

func (api *AppService) ScaleDown(request *restful.Request, response *restful.Response) {
	var param struct {
		RemoveInstances int `json:"removeInstances"`
	}

	err := request.ReadEntity(&param)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
	}

	err = api.Scheduler.ScaleDown(request.PathParameter("app_id"), param.RemoveInstances)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
	} else {
		response.WriteHeader(http.StatusOK)
	}
}

func (api *AppService) ScaleUp(request *restful.Request, response *restful.Response) {
	var param struct {
		NewInstances int      `json:"instances"`
		Ip           []string `json:"ip"`
	}
	err := request.ReadEntity(&param)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
	}

	err = api.Scheduler.ScaleUp(request.PathParameter("app_id"), param.NewInstances, param.Ip)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
	} else {
		response.WriteHeader(http.StatusOK)
	}
}

func (api *AppService) UpdateApp(request *restful.Request, response *restful.Response) {
	var version types.Version

	err := request.ReadEntity(&version)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
	}

	if CheckVersion(&version) == nil {
		err := api.Scheduler.UpdateApp(request.PathParameter("app_id"), &version)
		if err != nil {
			response.WriteError(http.StatusBadRequest, err)
		} else {
			response.WriteHeaderAndJson(http.StatusOK, []string{"version accepted"}, restful.MIME_JSON)
		}
	} else {
		response.WriteErrorString(http.StatusBadRequest, "Invalid Version.")
	}
}

func (api *AppService) ProceedUpdate(request *restful.Request, response *restful.Response) {
	var param struct {
		Instances int `json:"instances"`
	}

	err := request.ReadEntity(&param)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
	}

	err = api.Scheduler.ProceedUpdate(request.PathParameter("app_id"), param.Instances)
	if err != nil {
		logrus.Errorf("%s", err)
		response.WriteError(http.StatusBadRequest, err)
	} else {
		response.WriteHeaderAndJson(http.StatusOK, []string{"version accepted"}, restful.MIME_JSON)
	}
}

func (api *AppService) CancelUpdate(request *restful.Request, response *restful.Response) {
	err := api.Scheduler.CancelUpdate(request.PathParameter("app_id"))
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
	} else {
		response.WriteHeaderAndJson(http.StatusOK, []string{"version accepted"}, restful.MIME_JSON)
	}
}

func (api *Api) Start(ctx context.Context) error {
	router := gin.Default()

	group := router.Group(API_PREFIX)
	group.GET("/apps", api.ListApp)
	group.POST("/apps", api.CreateApp)
	group.GET("/apps/:app_id", api.GetApp)
	group.DELETE("/apps/:app_id", api.DeleteApp)
	group.PUT("/apps/:app_id/scale_up", api.ScaleUp)
	group.PUT("/apps/:app_id/scale_down", api.ScaleDown)
	group.PUT("/apps/:app_id/update", api.UpdateApp)
	group.PUT("/apps/:app_id/proceed_update", api.ProceedUpdate)
	group.PUT("/apps/:app_id/cancel_update", api.CancelUpdate)

	group.GET("/apps/:app_id/tasks", api.GetApp)
	group.DELETE("/apps/:app_id/tasks", api.GetApp) // pending
	group.DELETE("/apps/:app_id/tasks/:task_id", api.GetApp)

	group.GET("/apps/:app_id/versions", api.GetApp)
	group.GET("/apps/:app_id/versions/:version_id", api.GetApp)

	router.Run(api.config.HttpAddr)
	return nil
}

func (api *Api) CreateApp(c *gin.Context) {
	var version types.Version

	if c.BindJSON(&version) == nil && CheckVersion(&version) == nil {
		_, err := api.Scheduler.CreateApp(&version)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "version accepted"})
		}
	} else {
		c.JSON(400, gin.H{"status": "unauthorized"})
	}
}

func (api *Api) ListApp(c *gin.Context) {
	appsRet := make([]*App, 0)
	for _, app := range api.Scheduler.ListApps() {
		version := app.CurrentVersion
		appsRet = append(appsRet, &App{
			ID:               version.AppId,
			Name:             version.AppId,
			Instances:        int(version.Instances),
			RunningInstances: app.RunningInstances(),
			RunAs:            version.RunAs,
			ClusterId:        app.MesosConnector.ClusterId,
			Created:          app.Created,
			Updated:          app.Updated,
			Mode:             string(app.Mode),
			State:            app.State,
		})
	}

	c.JSON(http.StatusOK, gin.H{"apps": appsRet})
}

func (api *Api) GetApp(c *gin.Context) {
	app, err := api.Scheduler.InspectApp(c.Param("app_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	} else {
		version := app.CurrentVersion
		appRet := &App{
			ID:               version.AppId,
			Name:             version.AppId,
			Instances:        int(version.Instances),
			RunningInstances: app.RunningInstances(),
			RunAs:            version.RunAs,
			ClusterId:        app.MesosConnector.ClusterId,
			Created:          app.Created,
			Updated:          app.Updated,
			Mode:             string(app.Mode),
			State:            app.State,
		}

		appRet.Versions = make([]string, 0)
		for _, v := range app.Versions {
			appRet.Versions = append(appRet.Versions, v.ID)
		}

		appRet.Tasks = FilterTasksFromApp(app)

		c.JSON(http.StatusOK, gin.H{"app": appRet})
	}
}

func (api *Api) ScaleDown(c *gin.Context) {
	var param struct {
		RemoveInstances int `json:"removeInstances"`
	}

	if c.BindJSON(&param) != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	err := api.Scheduler.ScaleDown(c.Param("app_id"), param.RemoveInstances)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}

func (api *Api) ScaleUp(c *gin.Context) {
	var param struct {
		NewInstances int      `json:"instances"`
		Ip           []string `json:"ip"`
	}
	if c.BindJSON(&param) != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	err := api.Scheduler.ScaleUp(c.Param("app_id"), param.NewInstances, param.Ip)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}

func (api *Api) DeleteApp(c *gin.Context) {
	err := api.Scheduler.DeleteApp(c.Param("app_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}

func (api *Api) UpdateApp(c *gin.Context) {
	var version types.Version

	if c.BindJSON(&version) == nil && CheckVersion(&version) == nil {
		err := api.Scheduler.UpdateApp(c.Param("app_id"), &version)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "version accepted"})
		}
	} else {
		c.JSON(400, gin.H{"status": "unauthorized"})
	}
}

func (api *Api) ProceedUpdate(c *gin.Context) {
	var param struct {
		Instances int `json:"instances"`
	}

	if c.BindJSON(&param) == nil {
		err := api.Scheduler.ProceedUpdate(c.Param("app_id"), param.Instances)
		if err != nil {
			logrus.Errorf("%s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "version accepted"})
		}
	} else {
		c.JSON(400, gin.H{"status": "unauthorized"})
	}
}

func (api *Api) CancelUpdate(c *gin.Context) {
	err := api.Scheduler.CancelUpdate(c.Param("app_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "version accepted"})
	}
}

func CheckVersion(version *types.Version) error {
	// image format
	// mode valid
	// instance exists
	return nil
}

func FilterTasksFromApp(app *state.App) []*Task {
	tasks := make([]*Task, 0)
	for _, slot := range app.Slots {
		task := &Task{ // aka Slot
			ID:            slot.Id,
			AppId:         slot.App.AppId, // either Name or Id, rename AppId later
			VersionId:     slot.Version.ID,
			Status:        string(slot.State),
			OfferId:       slot.OfferId,
			AgentId:       slot.AgentId,
			AgentHostname: slot.AgentHostName,
			History:       make([]*TaskHistory, 0), // aka Task
			Cpu:           slot.Version.Cpus,
			Mem:           slot.Version.Mem,
			Disk:          slot.Version.Disk,
		}

		if len(slot.TaskHistory) > 0 {
			for _, v := range slot.TaskHistory {
				staleTask := &TaskHistory{
					ID:            v.Id,
					State:         v.State,
					Reason:        v.Reason,
					OfferId:       v.OfferId,
					AgentId:       v.AgentId,
					AgentHostname: v.AgentHostName,
					VersionId:     v.Version.ID,

					Cpu:  v.Version.Cpus,
					Mem:  v.Version.Mem,
					Disk: v.Version.Disk,

					Stderr: v.Stderr,
					Stdout: v.Stdout,
				}

				task.History = append(task.History, staleTask)
			}
		}

		tasks = append(tasks, task)
	}

	return tasks
}
