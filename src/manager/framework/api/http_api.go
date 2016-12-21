package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
)

const (
	API_PREFIX = "v_beta"
)

type AppService struct {
	Scheduler *scheduler.Scheduler
	apiserver.ApiRegister
}

func NewAndInstallAppService(apiServer *apiserver.ApiServer, eng *scheduler.Scheduler) *AppService {
	appService := &AppService{
		Scheduler: eng,
	}
	apiserver.Install(apiServer, appService)
	return appService
}

// NOTE(xychu): Every service need to registed to ApiServer need to impl
//              a `Register` interface so that it can be added to ApiServer.Start
func (api *AppService) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		ApiVersion(API_PREFIX).
		Path("/" + API_PREFIX + "/apps").
		Doc("App management").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(api.ListApp).
		// docs
		Doc("List Apps").
		Operation("listApps").
		Returns(200, "OK", []App{}))
	ws.Route(ws.POST("/").To(api.CreateApp).
		// docs
		Doc("Create App").
		Operation("createApp").
		Returns(201, "OK", App{}).
		Returns(400, "BadRequest", nil).
		Reads(types.Version{}).
		Writes(App{}))
	ws.Route(ws.GET("/{app_id}").To(api.GetApp).
		// docs
		Doc("Get an App").
		Operation("getApp").
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")).
		Returns(200, "OK", App{}).
		Returns(404, "NotFound", nil).
		Writes(App{}))
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
		Returns(400, "BadRequest", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PATCH("/{app_id}/scale-down").To(api.ScaleDown).
		// docs
		Doc("Scale Down App").
		Operation("scaleDown").
		Returns(400, "BadRequest", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PUT("/{app_id}").To(api.UpdateApp).
		// docs
		Doc("Update App").
		Operation("updateApp").
		Returns(404, "NotFound", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PATCH("/{app_id}/proceed-update").To(api.ProceedUpdate).
		// docs
		Doc("Proceed Update App").
		Operation("proceedUpdateApp").
		Returns(400, "BadRequest", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PATCH("/{app_id}/cancel-update").To(api.CancelUpdate).
		// docs
		Doc("Cancel Update App").
		Operation("cancelUpdateApp").
		Returns(400, "BadRequest", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))

	ws.Route(ws.GET("/{app_id}/tasks/{task_id}").To(api.GetAppTask).
		// docs
		Doc("Get a task in the given App").
		Operation("getAppTask").
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")).
		Param(ws.PathParameter("task_id", "identifier of the task").DataType("int")).
		Returns(200, "OK", Task{}).
		Returns(404, "NotFound", nil))

	container.Add(ws)
}

func (api *AppService) CreateApp(request *restful.Request, response *restful.Response) {
	var version types.Version

	err := request.ReadEntity(&version)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, err.Error())
	}

	err = CheckVersion(&version)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, err.Error())
	}

	app, err := api.Scheduler.CreateApp(&version)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	} else {
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
			Labels:           version.Labels,
			Env:              version.Env,
			Constraints:      version.Constraints,
			Uris:             version.Uris,
		}

		appRet.Versions = make([]string, 0)
		for _, v := range app.Versions {
			appRet.Versions = append(appRet.Versions, v.ID)
		}

		appRet.Tasks = FilterTasksFromApp(app)
		response.WriteHeaderAndEntity(http.StatusCreated, appRet)
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
			Labels:           version.Labels,
			Env:              version.Env,
			Constraints:      version.Constraints,
			Uris:             version.Uris,
		})
	}

	response.WriteEntity(appsRet)
}

func (api *AppService) GetApp(request *restful.Request, response *restful.Response) {
	app, err := api.Scheduler.InspectApp(request.PathParameter("app_id"))
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, err.Error())
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
			Labels:           version.Labels,
			Env:              version.Env,
			Constraints:      version.Constraints,
			Uris:             version.Uris,
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
		response.WriteErrorString(http.StatusNotFound, err.Error())
	} else {
		response.WriteHeader(http.StatusNoContent)
	}
}

func (api *AppService) ScaleDown(request *restful.Request, response *restful.Response) {
	var param struct {
		RemoveInstances int `json:"instances"`
	}

	err := request.ReadEntity(&param)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, err.Error())
	}

	err = api.Scheduler.ScaleDown(request.PathParameter("app_id"), param.RemoveInstances)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, err.Error())
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
		response.WriteErrorString(http.StatusBadRequest, err.Error())
	}

	err = api.Scheduler.ScaleUp(request.PathParameter("app_id"), param.NewInstances, param.Ip)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, err.Error())
	} else {
		response.WriteHeader(http.StatusOK)
	}
}

func (api *AppService) UpdateApp(request *restful.Request, response *restful.Response) {
	var version types.Version

	err := request.ReadEntity(&version)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, err.Error())
	}

	if CheckVersion(&version) == nil {
		err := api.Scheduler.UpdateApp(request.PathParameter("app_id"), &version)
		if err != nil {
			response.WriteErrorString(http.StatusBadRequest, err.Error())
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
		response.WriteErrorString(http.StatusBadRequest, err.Error())
	}

	err = api.Scheduler.ProceedUpdate(request.PathParameter("app_id"), param.Instances)
	if err != nil {
		logrus.Errorf("%s", err)
		response.WriteErrorString(http.StatusBadRequest, err.Error())
	} else {
		response.WriteHeaderAndJson(http.StatusOK, []string{"version accepted"}, restful.MIME_JSON)
	}
}

func (api *AppService) CancelUpdate(request *restful.Request, response *restful.Response) {
	err := api.Scheduler.CancelUpdate(request.PathParameter("app_id"))
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, err.Error())
	} else {
		response.WriteHeaderAndJson(http.StatusOK, []string{"version accepted"}, restful.MIME_JSON)
	}
}

func (api *AppService) GetAppTask(request *restful.Request, response *restful.Response) {
	app, err := api.Scheduler.InspectApp(request.PathParameter("app_id"))
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, err.Error())
	} else {
		task_id := request.PathParameter("task_id")
		task_index, err := strconv.Atoi(task_id)
		if err != nil {
			response.WriteErrorString(http.StatusBadRequest, "Get task index err: "+err.Error())
		} else {
			appTaskRet, err := GetTaskFromApp(app, task_index)
			if err != nil {
				response.WriteErrorString(http.StatusBadRequest, "Get task err: "+err.Error())
			} else {
				response.WriteEntity(appTaskRet)
			}
		}
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
			Created:       slot.CurrentTask.Created,
			Image:         slot.Version.Container.Docker.Image,
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

func GetTaskFromApp(app *state.App, task_index int) (*Task, error) {
	slot, found := app.Slots[task_index]
	if !found {
		logrus.Errorf("slot not found: %s", task_index)
		return nil, errors.New("slot not found")
	}

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
		Created:       slot.CurrentTask.Created,
		Image:         slot.Version.Container.Docker.Image,
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

	return task, nil
}
