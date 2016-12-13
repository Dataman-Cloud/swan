package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

const (
	API_PREFIX = "v_try"
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
		err := api.Scheduler.CreateApp(&version)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			IP:               version.Ip,
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
			IP:            slot.Ip,
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
