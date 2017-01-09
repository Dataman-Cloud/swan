package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Dataman-Cloud/swan/src/apiserver"
	"github.com/Dataman-Cloud/swan/src/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/utils/fields"
	"github.com/Dataman-Cloud/swan/src/utils/labels"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
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
		ApiVersion(config.API_PREFIX).
		Path(config.API_PREFIX + "/apps").
		Doc("App management").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(metrics.InstrumentRouteFunc("GET", "Apps", api.ListApp)).
		// docs
		Doc("List Apps").
		Operation("listApps").
		Param(ws.QueryParameter("labels", "app labels, e.g. labels=USER==1,cluster=beijing").DataType("string")).
		Param(ws.QueryParameter("fields", "app fields, e.g. runAs==xxx").DataType("string")).
		Returns(200, "OK", []types.App{}))
	ws.Route(ws.POST("/").To(metrics.InstrumentRouteFunc("POST", "App", api.CreateApp)).
		// docs
		Doc("Create App").
		Operation("createApp").
		Returns(201, "OK", types.App{}).
		Returns(400, "BadRequest", nil).
		Reads(types.Version{}).
		Writes(types.App{}))
	ws.Route(ws.GET("/{app_id}").To(metrics.InstrumentRouteFunc("GET", "App", api.GetApp)).
		// docs
		Doc("Get an App").
		Operation("getApp").
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")).
		Returns(200, "OK", types.App{}).
		Returns(404, "NotFound", nil).
		Writes(types.App{}))
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
		Reads(types.ScaleUpParam{}).
		Returns(200, "OK", nil).
		Returns(400, "BadRequest", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PATCH("/{app_id}/scale-down").To(metrics.InstrumentRouteFunc("PATCH", "App", api.ScaleDown)).
		// docs
		Doc("Scale Down App").
		Operation("scaleDown").
		Reads(types.ScaleDownParam{}).
		Returns(200, "OK", nil).
		Returns(400, "BadRequest", nil).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PUT("/{app_id}").To(metrics.InstrumentRouteFunc("PUT", "App", api.UpdateApp)).
		// docs
		Doc("Update App").
		Operation("updateApp").
		Returns(200, "OK", types.App{}).
		Returns(404, "NotFound", nil).
		Reads(types.Version{}).
		Writes(types.App{}).
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")))
	ws.Route(ws.PATCH("/{app_id}/proceed-update").To(metrics.InstrumentRouteFunc("PATCH", "App", api.ProceedUpdate)).
		// docs
		Doc("Proceed Update App").
		Operation("proceedUpdateApp").
		Returns(400, "BadRequest", nil).
		Reads(types.ProceedUpdateParam{}).
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
		Returns(200, "OK", types.Task{}).
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

	response.WriteHeaderAndEntity(http.StatusCreated, FormAppRetWithVersionsAndTasks(app))
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

	appsRet := make([]*types.App, 0)
	for _, app := range api.Scheduler.ListApps(appFilterOptions) {
		appsRet = append(appsRet, FormAppRet(app))
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
	response.WriteEntity(FormAppRetWithVersionsAndTasks(app))
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
	var param types.ScaleDownParam

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
	var param types.ScaleUpParam
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
		appID := request.PathParameter("app_id")
		err := api.Scheduler.UpdateApp(appID, &version)
		if err != nil {
			logrus.Errorf("Update app[%s] error: %s", appID, err.Error())
			response.WriteError(http.StatusBadRequest, err)
			return
		}
		app, err := api.Scheduler.InspectApp(appID)
		if err != nil {
			logrus.Errorf("Inspect app[%s] error: %s", appID, err.Error())
			response.WriteError(http.StatusNotFound, err)
			return
		}
		response.WriteEntity(FormAppRetWithVersionsAndTasks(app))
	} else {
		response.WriteErrorString(http.StatusBadRequest, "Invalid Version.")
	}
}

func (api *AppService) ProceedUpdate(request *restful.Request, response *restful.Response) {
	var param types.ProceedUpdateParam

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
	versionID := request.PathParameter("version_id")

	if versionID == app.CurrentVersion.ID {
		response.WriteEntity(app.CurrentVersion)
		return
	}

	for _, v := range app.Versions {
		if v.ID == versionID {
			response.WriteEntity(v)
			return
		}
	}
	logrus.Errorf("No versions found with ID: %s", versionID)
	response.WriteErrorString(http.StatusNotFound, "No versions found")
}

func FormAppRet(app *state.App) *types.App {
	version := app.CurrentVersion
	appRet := &types.App{
		ID:               app.ID,
		Name:             app.Name,
		Instances:        int(version.Instances),
		RunningInstances: app.RunningInstances(),
		RunAs:            version.RunAs,
		Priority:         int(version.Priority),
		ClusterID:        app.ClusterID,
		Created:          app.Created,
		Updated:          app.Updated,
		Mode:             string(app.Mode),
		State:            app.State,
		CurrentVersion:   app.CurrentVersion,
		ProposedVersion:  app.ProposedVersion,
		Labels:           version.Labels,
		Env:              version.Env,
		Constraints:      version.Constraints,
		URIs:             version.URIs,
	}
	return appRet
}

func FormAppRetWithVersions(app *state.App) *types.App {
	appRet := FormAppRet(app)
	appRet.Versions = make([]string, 0)
	for _, v := range app.Versions {
		appRet.Versions = append(appRet.Versions, v.ID)
	}
	return appRet
}

func FormAppRetWithVersionsAndTasks(app *state.App) *types.App {
	appRet := FormAppRetWithVersions(app)
	appRet.Tasks = FilterTasksFromApp(app)
	return appRet
}

func CheckVersion(version *types.Version) error {
	// image format
	// mode valid
	// instance exists
	return nil
}

func FilterTasksFromApp(app *state.App) []*types.Task {
	tasks := make([]*types.Task, 0)
	for _, slot := range app.GetSlots() {
		task := &types.Task{ // aka Slot
			ID:            slot.ID,
			AppID:         slot.App.ID, // either Name or ID
			VersionID:     slot.Version.ID,
			Healthy:       slot.Healthy(),
			Status:        string(slot.State),
			OfferID:       slot.OfferID,
			AgentID:       slot.AgentID,
			AgentHostname: slot.AgentHostName,
			History:       make([]*types.TaskHistory, 0), // aka Task
			CPU:           slot.Version.CPUs,
			Mem:           slot.Version.Mem,
			Disk:          slot.Version.Disk,
			IP:            slot.Ip,
			Created:       slot.CurrentTask.Created,
			Image:         slot.Version.Container.Docker.Image,
		}

		if len(slot.TaskHistory) > 0 {
			for _, v := range slot.TaskHistory {
				staleTask := &types.TaskHistory{
					ID:            v.ID,
					State:         v.State,
					Reason:        v.Reason,
					OfferID:       v.OfferID,
					AgentID:       v.AgentID,
					AgentHostname: v.AgentHostName,

					Stderr: v.Stderr,
					Stdout: v.Stdout,
				}
				if v.Version != nil {
					staleTask.VersionID = v.Version.ID
					staleTask.CPU = v.Version.CPUs
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

func GetTaskFromApp(app *state.App, task_index int) (*types.Task, error) {
	slots := app.GetSlots()
	if task_index > len(slots)-1 || task_index < 0 {
		logrus.Errorf("slot not found: %d", task_index)
		return nil, errors.New("slot task found")
	}

	slot := slots[task_index]

	task := &types.Task{ // aka Slot
		ID:            slot.ID,
		AppID:         slot.App.ID, // either Name or ID
		VersionID:     slot.Version.ID,
		Status:        string(slot.State),
		OfferID:       slot.OfferID,
		AgentID:       slot.AgentID,
		AgentHostname: slot.AgentHostName,
		History:       make([]*types.TaskHistory, 0), // aka Task
		CPU:           slot.Version.CPUs,
		Mem:           slot.Version.Mem,
		Disk:          slot.Version.Disk,
		IP:            slot.Ip,
		Created:       slot.CurrentTask.Created,
		Image:         slot.Version.Container.Docker.Image,
	}

	if len(slot.TaskHistory) > 0 {
		for _, v := range slot.TaskHistory {
			if v == nil {
				continue
			}

			task.History = append(task.History, FormTaskRet(v))
		}
	}

	if slot.CurrentTask != nil {
		task.CurrentTask = FormTaskRet(slot.CurrentTask)
	}

	return task, nil
}

func FormTaskRet(v *state.Task) *types.TaskHistory {
	return &types.TaskHistory{
		ID:            v.ID,
		State:         v.State,
		Reason:        v.Reason,
		OfferID:       v.OfferID,
		AgentID:       v.AgentID,
		AgentHostname: v.AgentHostName,
		VersionID:     v.Version.ID,

		CPU:  v.Version.CPUs,
		Mem:  v.Version.Mem,
		Disk: v.Version.Disk,

		Stderr: v.Stderr,
		Stdout: v.Stdout,
	}

}
