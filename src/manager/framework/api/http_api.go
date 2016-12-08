package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	//"github.com/emicklei/go-restful/swagger"
)

const (
	API_PREFIX = "v_try"
)

type Api struct {
	port      string
	Scheduler *scheduler.Scheduler
}

func NewApi(eng *scheduler.Scheduler) *Api {
	return &Api{
		port:      ":12306",
		Scheduler: eng,
	}
}

type AppService struct {
	addr      string
	Scheduler *scheduler.Scheduler
}

func NewAppService(eng *scheduler.Scheduler) *AppService {
	return &AppService{
		addr:      ":8080",
		Scheduler: eng,
	}
}

func (api *AppService) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	//ws.Path(API_PREFIX)

	ws.Route(ws.GET("/apps").To(api.ListApp).
		// docs
		Doc("List Apps").
		Operation("listApps").
		Returns(200, "OK", []types.Application{}))

	ws.Route(ws.POST("/apps").To(api.CreateApp).
		// docs
		Doc("Create App").
		Operation("createApp").
		Returns(201, "OK", []types.Application{}).
		Writes(types.Application{}))
	//ws.Route(ws.GET("/apps/{app_id}").To(api.GetApp).
	//	// docs
	//	Doc("Get an App").
	//	Operation("getApp").
	//	Returns(200, "OK", types.Application{}).
	//	Writes(types.Application{}))
	//ws.Route(ws.DELETE("/apps/{app_id}").To(api.DeleteApp).
	//	// docs
	//	Doc("Delete App").
	//	Operation("deleteApp"))

	container.Add(ws)
}

// TODO(xychu): Will move to a global place later
func (api *AppService) Start() error {
	wsContainer := restful.NewContainer()

	api.Register(wsContainer)

	//// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	//// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	//// Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
	//config := swagger.Config{
	//	WebServices:    wsContainer.RegisteredWebServices(), // you control what services are visible
	//	WebServicesUrl: "http://localhost:8080",
	//	ApiPath:        "/apidocs.json",
	//
	//	// Optionally, specifiy where the UI is located
	//	SwaggerPath:     "/apidocs/",
	//	SwaggerFilePath: "/Users/chuxiangyang/go/src/github.com/Dataman-Cloud/swan/example/dist",
	//	}
	//swagger.RegisterSwaggerService(config, wsContainer)

	logrus.Printf("start listening on %s", api.addr)
	server := &http.Server{Addr: api.addr, Handler: wsContainer}
	logrus.Fatal(server.ListenAndServe())

	return nil
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
		response.WriteEntity(app)
	}
}

func (api *AppService) ListApp(request *restful.Request, response *restful.Response) {
	appsRet := make([]*App, 0)
	for _, app := range api.Scheduler.ListApps() {
		version := app.CurrentVersion
		appsRet = append(appsRet, &App{
			ID:                version.AppId,
			Name:              version.AppId,
			Instances:         int(version.Instances),
			RunningInstances:  app.RunningInstances(),
			RollbackInstances: app.RollbackInstances(),
			RunAs:             version.RunAs,
			ClusterId:         app.MesosConnector.ClusterId,
			Created:           app.Created,
			Updated:           app.Updated,
			Mode:              string(app.Mode),
			State:             app.State,
		})
	}

	response.WriteEntity(appsRet)
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

	router.Run(api.port)
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
		To int `json:"to"`
	}

	if c.BindJSON(&param) != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	err := api.Scheduler.ScaleDown(c.Param("app_id"), param.To)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}

func (api *Api) ScaleUp(c *gin.Context) {
	var param struct {
		To int `json:"to"`
	}
	if c.BindJSON(&param) != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	err := api.Scheduler.ScaleUp(c.Param("app_id"), param.To)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
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