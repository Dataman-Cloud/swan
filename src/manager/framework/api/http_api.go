package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/utils/fields"
	"github.com/Dataman-Cloud/swan/src/utils/labels"

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

	ws.Route(ws.GET("/").To(metrics.InstrumentRouteFunc("GET", "Apps", api.ListApp)).
		// docs
		Doc("List Apps").
		Operation("listApps").
		Param(ws.QueryParameter("labels", "app labels, e.g. labels=USER==1,cluster=beijing").DataType("string")).
		Param(ws.QueryParameter("fields", "app fields, e.g. runAs==xxx").DataType("string")).
		Returns(200, "OK", []App{}))
	ws.Route(ws.POST("/").To(metrics.InstrumentRouteFunc("POST", "App", api.CreateApp)).
		// docs
		Doc("Create App").
		Operation("createApp").
		Returns(201, "OK", App{}).
		Returns(400, "BadRequest", nil).
		Reads(types.Version{}).
		Writes(App{}))
	ws.Route(ws.GET("/{app_id}").To(metrics.InstrumentRouteFunc("GET", "App", api.GetApp)).
		// docs
		Doc("Get an App").
		Operation("getApp").
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")).
		Returns(200, "OK", App{}).
		Returns(404, "NotFound", nil).
		Writes(App{}))
	ws.Route(ws.DELETE("/{app_id}").To(metrics.InstrumentRouteFunc("DELETE", "App", api.DeleteApp)).
		// docs
		Doc("Delete App").
		Operation("deleteApp").
		Returns(204, "OK", nil).
		Returns(404, "NotFound", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PATCH("/{app_id}/scale-up").To(metrics.InstrumentRouteFunc("PATCH", "App", api.ScaleUp)).
		// docs
		Doc("Scale Up App").
		Operation("scaleUp").
		Reads(ScaleUpParam{}).
		Returns(200, "OK", nil).
		Returns(400, "BadRequest", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PATCH("/{app_id}/scale-down").To(metrics.InstrumentRouteFunc("PATCH", "App", api.ScaleDown)).
		// docs
		Doc("Scale Down App").
		Operation("scaleDown").
		Reads(ScaleDownParam{}).
		Returns(200, "OK", nil).
		Returns(400, "BadRequest", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PUT("/{app_id}").To(metrics.InstrumentRouteFunc("PUT", "App", api.UpdateApp)).
		// docs
		Doc("Update App").
		Operation("updateApp").
		Returns(200, "OK", App{}).
		Returns(404, "NotFound", nil).
		Reads(types.Version{}).
		Writes(App{}).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PATCH("/{app_id}/proceed-update").To(metrics.InstrumentRouteFunc("PATCH", "App", api.ProceedUpdate)).
		// docs
		Doc("Proceed Update App").
		Operation("proceedUpdateApp").
		Returns(400, "BadRequest", nil).
		Reads(ProceedUpdateParam{}).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PATCH("/{app_id}/cancel-update").To(metrics.InstrumentRouteFunc("PATCH", "App", api.CancelUpdate)).
		// docs
		Doc("Cancel Update App").
		Operation("cancelUpdateApp").
		Returns(200, "OK", nil).
		Returns(400, "BadRequest", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))

	ws.Route(ws.GET("/{app_id}/tasks/{task_id}").To(metrics.InstrumentRouteFunc("GET", "AppTask", api.GetAppTask)).
		// docs
		Doc("Get a task in the given App").
		Operation("getAppTask").
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")).
		Param(ws.PathParameter("task_id", "identifier of the task").DataType("int")).
		Returns(200, "OK", Task{}).
		Returns(404, "NotFound", nil))

	ws.Route(ws.GET("/{app_id}/versions").To(metrics.InstrumentRouteFunc("GET", "AppVersions", api.GetAppVersions)).
		// docs
		Doc("Get all versions in the given App").
		Operation("getAppVersions").
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")).
		Returns(200, "OK", []types.Version{}).
		Returns(404, "NotFound", nil))

	ws.Route(ws.GET("/{app_id}/versions/{version_id}").To(metrics.InstrumentRouteFunc("GET", "AppVersion", api.GetAppVersion)).
		// docs
		Doc("Get a version in the given App").
		Operation("getAppVersion").
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")).
		Param(ws.PathParameter("version_id", "identifier of the app version").DataType("string")).
		Returns(200, "OK", types.Version{}).
		Returns(404, "NotFound", nil))

	container.Add(ws)
}

func (api *AppService) CreateApp(request *restful.Request, response *restful.Response) {
	var version types.Version

	err := request.ReadEntity(&version)
	if err != nil {
		logrus.Errorf("Create app error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	err = CheckVersion(&version)
	if err != nil {
		logrus.Errorf("Create app error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	app, err := api.Scheduler.CreateApp(&version)
	if err != nil {
		logrus.Errorf("Create app error: %s", err.Error())
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	appRet := &App{
		ID:               version.AppId,
		Name:             version.AppId,
		Instances:        int(version.Instances),
		RunningInstances: app.RunningInstances(),
		RunAs:            version.RunAs,
		ClusterId:        app.ClusterId,
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

func (api *AppService) ListApp(request *restful.Request, response *restful.Response) {
	appFilterOptions := scheduler.AppFilterOptions{}
	labelsFilter := request.QueryParameter("labels")
	if labelsFilter != "" {
		labelsSelector, err := labels.Parse(labelsFilter)
		if err != nil {
			logrus.Errorf("parse condition of label %s failed. Error: %+v", labelsSelector, err)
		} else {
			appFilterOptions.LabelsSelector = labelsSelector
		}
	}

	fieldsFilter := request.QueryParameter("fields")
	if fieldsFilter != "" {
		fieldSelector, err := fields.ParseSelector(fieldsFilter)
		if err != nil {
			logrus.Errorf("parse condition of field %s failed. Error: %+v", fieldsFilter, err)
		} else {
			appFilterOptions.FieldsSelector = fieldSelector
		}
	}

	appsRet := make([]*App, 0)
	for _, app := range api.Scheduler.ListApps(appFilterOptions) {
		version := app.CurrentVersion
		appsRet = append(appsRet, &App{
			ID:               version.AppId,
			Name:             version.AppId,
			Instances:        int(version.Instances),
			RunningInstances: app.RunningInstances(),
			RunAs:            version.RunAs,
			ClusterId:        app.ClusterId,
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
		logrus.Errorf("Get app error: %s", err.Error())
		response.WriteError(http.StatusNotFound, err)
		return
	}
	response.WriteEntity(FormAppRet(app))
}

func (api *AppService) DeleteApp(request *restful.Request, response *restful.Response) {
	err := api.Scheduler.DeleteApp(request.PathParameter("app_id"))
	if err != nil {
		logrus.Errorf("Delete app error: %s", err.Error())
		response.WriteError(http.StatusNotFound, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (api *AppService) ScaleDown(request *restful.Request, response *restful.Response) {
	var param ScaleDownParam

	err := request.ReadEntity(&param)
	if err != nil {
		logrus.Errorf("Scale down app error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
	}

	err = api.Scheduler.ScaleDown(request.PathParameter("app_id"), param.Instances)
	if err != nil {
		logrus.Errorf("Scale down app error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (api *AppService) ScaleUp(request *restful.Request, response *restful.Response) {
	var param ScaleUpParam
	err := request.ReadEntity(&param)
	if err != nil {
		logrus.Errorf("Scale up app error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	err = api.Scheduler.ScaleUp(request.PathParameter("app_id"), param.Instances, param.IPs)
	if err != nil {
		logrus.Errorf("Scale up app error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (api *AppService) UpdateApp(request *restful.Request, response *restful.Response) {
	var version types.Version

	err := request.ReadEntity(&version)
	if err != nil {
		logrus.Errorf("Update app error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	if CheckVersion(&version) == nil {
		err := api.Scheduler.UpdateApp(request.PathParameter("app_id"), &version)
		if err != nil {
			logrus.Errorf("Update app error: %s", err.Error())
			response.WriteError(http.StatusBadRequest, err)
			return
		}
		app, err := api.Scheduler.InspectApp(request.PathParameter("app_id"))
		if err != nil {
			logrus.Errorf("Update app error: %s", err.Error())
			response.WriteError(http.StatusNotFound, err)
			return
		}
		response.WriteEntity(FormAppRet(app))
	} else {
		response.WriteErrorString(http.StatusBadRequest, "Invalid Version.")
	}
}

func (api *AppService) ProceedUpdate(request *restful.Request, response *restful.Response) {
	var param ProceedUpdateParam

	err := request.ReadEntity(&param)
	if err != nil {
		logrus.Errorf("Proceed update app error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	err = api.Scheduler.ProceedUpdate(request.PathParameter("app_id"), param.Instances)
	if err != nil {
		logrus.Errorf("Proceed update error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	response.Write([]byte("Update proceeded"))
}

func (api *AppService) CancelUpdate(request *restful.Request, response *restful.Response) {
	err := api.Scheduler.CancelUpdate(request.PathParameter("app_id"))
	if err != nil {
		logrus.Errorf("Cancel update error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	response.Write([]byte("Update canceled"))
}

func (api *AppService) GetAppTask(request *restful.Request, response *restful.Response) {
	app, err := api.Scheduler.InspectApp(request.PathParameter("app_id"))
	if err != nil {
		logrus.Errorf("Get app task error: %s", err.Error())
		response.WriteError(http.StatusNotFound, err)
		return
	}
	task_id := request.PathParameter("task_id")
	task_index, err := strconv.Atoi(task_id)
	if err != nil {
		logrus.Errorf("Get task index err: %s", err.Error())
		response.WriteErrorString(http.StatusBadRequest, "Get task index err: "+err.Error())
		return
	}

	appTaskRet, err := GetTaskFromApp(app, task_index)
	if err != nil {
		logrus.Errorf("Get task err: %s", err.Error())
		response.WriteErrorString(http.StatusBadRequest, "Get task err: "+err.Error())
		return
	}
	response.WriteEntity(appTaskRet)
}

func (api *AppService) GetAppVersions(request *restful.Request, response *restful.Response) {
	app, err := api.Scheduler.InspectApp(request.PathParameter("app_id"))
	if err != nil {
		logrus.Errorf("Get app versions error: %s", err.Error())
		response.WriteError(http.StatusNotFound, err)
		return
	}
	response.WriteEntity(app.Versions)
}

func (api *AppService) GetAppVersion(request *restful.Request, response *restful.Response) {
	app, err := api.Scheduler.InspectApp(request.PathParameter("app_id"))
	if err != nil {
		logrus.Errorf("Get app versions error: %s", err.Error())
		response.WriteError(http.StatusNotFound, err)
		return
	}
	version_id := request.PathParameter("version_id")

	for _, v := range app.Versions {
		if v.ID == version_id {
			response.WriteEntity(v)
			return
		}
	}
	logrus.Errorf("No versions found with ID: %s", version_id)
	response.WriteErrorString(http.StatusNotFound, "No versions found")
}

func FormAppRet(app *state.App) *App {
	version := app.CurrentVersion
	appRet := &App{
		ID:               version.AppId,
		Name:             version.AppId,
		Instances:        int(version.Instances),
		RunningInstances: app.RunningInstances(),
		RunAs:            version.RunAs,
		ClusterId:        app.ClusterId,
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
	return appRet
}

func CheckVersion(version *types.Version) error {
	// image format
	// mode valid
	// instance exists
	return nil
}

func FilterTasksFromApp(app *state.App) []*Task {
	tasks := make([]*Task, 0)
	for _, slot := range app.GetSlots() {
		task := &Task{ // aka Slot
			ID:            slot.Id,
			AppId:         slot.App.AppId, // either Name or Id, rename AppId later
			VersionId:     slot.Version.ID,
			Healthy:       slot.Healthy(),
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

					Stderr: v.Stderr,
					Stdout: v.Stdout,
				}
				if v.Version != nil {
					staleTask.VersionId = v.Version.ID
					staleTask.Cpu = v.Version.Cpus
					staleTask.Mem = v.Version.Mem
					staleTask.Disk = v.Version.Disk
				}

				task.History = append(task.History, staleTask)
			}
		}

		tasks = append(tasks, task)
	}

	return tasks
}

func GetTaskFromApp(app *state.App, task_index int) (*Task, error) {
	slots := app.GetSlots()
	if task_index >= len(slots)-1 || task_index < 0 {
		logrus.Errorf("slot not found: %s", task_index)
		return nil, errors.New("slot task found")
	}

	slot := slots[task_index]

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
