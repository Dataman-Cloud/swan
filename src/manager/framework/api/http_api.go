package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/manager/framework/engine"
	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

const (
	API_PREFIX = "v_try"
)

type Api struct {
	port   string
	Engine *engine.Engine
}

func NewApi(eng *engine.Engine) *Api {
	return &Api{
		port:   ":12306",
		Engine: eng,
	}
}

func (api *Api) Start(ctx context.Context) error {
	router := gin.Default()

	group := router.Group(API_PREFIX)
	group.GET("/apps", api.ListApp)
	group.POST("/apps", api.CreateApp)
	group.GET("/apps/:app_id", api.GetApp)
	group.POST("/apps/:app_id/update", api.GetApp)
	group.POST("/apps/:app_id/scale", api.GetApp)
	group.POST("/apps/:app_id/rollback", api.GetApp)

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
		err := api.Engine.CreateApp(&version)
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
	for _, app := range api.Engine.ListApps() {
		version := app.CurrentVersion
		appsRet = append(appsRet, &App{
			ID:                version.AppId,
			Name:              version.AppId,
			Instances:         int(version.Instances),
			RunningInstances:  app.RunningInstances(),
			RollbackInstances: app.RollbackInstances(),
			RunAs:             version.RunAs,
			ClusterId:         app.ClusterId,
			Created:           app.Created,
			Updated:           app.Updated,
			Mode:              string(app.Mode),
		})
	}

	c.JSON(http.StatusOK, gin.H{"apps": appsRet})
}

func (api *Api) GetApp(c *gin.Context) {
	app, err := api.Engine.InspectApp(c.Param("app_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	} else {
		version := app.CurrentVersion
		appRet := &App{
			ID:                version.AppId,
			Name:              version.AppId,
			Instances:         int(version.Instances),
			RunningInstances:  app.RunningInstances(),
			RollbackInstances: app.RollbackInstances(),
			RunAs:             version.RunAs,
			ClusterId:         app.ClusterId,
			Created:           app.Created,
			Updated:           app.Updated,
			Mode:              string(app.Mode),
		}

		appRet.Versions = make([]string, 0)
		for _, v := range app.Versions {
			appRet.Versions = append(appRet.Versions, v.ID)
		}

		appRet.Tasks = FilterTasksFromApp(app)

		c.JSON(http.StatusOK, gin.H{"app": appRet})
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
	return tasks
}
