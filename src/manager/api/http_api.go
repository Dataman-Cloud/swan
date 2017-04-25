package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/manager/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/state"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/utils/fields"
	"github.com/Dataman-Cloud/swan/src/utils/labels"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
)

type AppService struct {
	Scheduler *scheduler.Scheduler
	apiServer *apiserver.ApiServer
}

func NewAndInstallAppService(apiServer *apiserver.ApiServer, eng *scheduler.Scheduler) {
	appService := &AppService{
		Scheduler: eng,
		apiServer: apiServer,
	}
	apiserver.Install(apiServer, appService)
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
	ws.Route(ws.PATCH("/{app_id}/weights").To(metrics.InstrumentRouteFunc("PATCH", "App", api.UpdateWeights)).
		// docs
		Doc("Update Slot Weight").
		Operation("updateWeight").
		Returns(400, "BadRequest", nil).
		Returns(200, "OK", types.App{}).
		Reads(types.UpdateWeightsParam{}).
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

	ws.Route(ws.PATCH("/{app_id}/tasks/{task_id}/weight").To(metrics.InstrumentRouteFunc("GET", "AppTask", api.UpdateAppTaskWeight)).
		// docs
		Doc("Update weight of a task").
		Operation("updateAppTaskWeight").
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

	ws.Route(ws.GET("/{app_id}/service-discoveries").To(metrics.InstrumentRouteFunc("GET", "AppServiceDiscoveries", api.GetAppServiceDiscoveries)).
		// docs
		Doc("Get an app's service discoveries").
		Operation("getAppServiceDiscoveries").
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")).
		Returns(200, "OK", []types.ServiceDiscovery{}).
		Returns(404, "NotFound", nil))

	ws.Route(ws.GET("/{app_id}/service-discoveries/md5").To(metrics.InstrumentRouteFunc("GET", "AppServiceDiscoveries", api.GetAppServiceDiscoveriesMD5)).
		// docs
		Doc("Get service discoveries' md5").
		Operation("getAppServiceDiscoveriesMD5").
		Param(ws.PathParameter("app_id", "identifier of the app").DataType("string")).
		Returns(200, "OK", "").
		Returns(404, "NotFound", nil))

	ws.Route(ws.GET("/service-discoveries").To(metrics.InstrumentRouteFunc("GET", "AppServiceDiscoveriesMD5", api.GetAllServiceDiscoveries)).
		// docs
		Doc("Get all apps' service discoveries").
		Operation("getAllServiceDiscoveries").
		Returns(200, "OK", []types.ServiceDiscovery{}).
		Returns(404, "NotFound", nil))

	ws.Route(ws.GET("/service-discoveries/md5").To(metrics.InstrumentRouteFunc("GET", "ServiceDiscoveriesMD5", api.GetAllServiceDiscoveriesMD5)).
		// docs
		Doc("Get all apps' service discoveries' md5").
		Operation("getAllServiceDiscoveriesMD5").
		Returns(200, "OK", "").
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

	app, err := api.Scheduler.CreateApp(&version)
	if err != nil {
		logrus.Errorf("Create app error: %s", err.Error())
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, FormAppRetWithVersions(app))
}

func (api *AppService) ListApp(request *restful.Request, response *restful.Response) {
	appFilterOptions := types.AppFilterOptions{}
	labelsFilter := request.QueryParameter("labels")
	if labelsFilter != "" {
		labelsSelector, err := labels.Parse(labelsFilter)
		if err != nil {
			logrus.Errorf("parse condition of label %s failed. Error: %+v", labelsSelector, err)
			response.WriteError(http.StatusBadRequest, err)
			return
		} else {
			appFilterOptions.LabelsSelector = labelsSelector
		}
	}

	fieldsFilter := request.QueryParameter("fields")
	if fieldsFilter != "" {
		fieldSelector, err := fields.ParseSelector(fieldsFilter)
		if err != nil {
			logrus.Errorf("parse condition of field %s failed. Error: %+v", fieldsFilter, err)
			response.WriteError(http.StatusBadRequest, err)
			return
		} else {
			appFilterOptions.FieldsSelector = fieldSelector
		}
	}

	appsRet := make([]*types.App, 0)
	for _, app := range api.Scheduler.ListApps(appFilterOptions) {
		appsRet = append(appsRet, FormAppWithTask(app))
	}

	response.WriteEntity(appsRet)
}

func (api *AppService) GetApp(request *restful.Request, response *restful.Response) {
	app, err := api.Scheduler.InspectApp(request.PathParameter("app_id"))
	if err != nil {
		logrus.Debugf("Get app error: %s", err.Error())
		response.WriteError(http.StatusNotFound, err)
		return
	}
	response.WriteEntity(FormAppRetWithVersions(app))
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

	appID := request.PathParameter("app_id")
	if err := api.Scheduler.UpdateApp(appID, &version); err != nil {
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
	response.WriteEntity(FormAppRetWithVersions(app))
	return
}

func (api *AppService) UpdateWeights(request *restful.Request, response *restful.Response) {
	var param types.UpdateWeightsParam

	err := request.ReadEntity(&param)
	if err != nil {
		logrus.Errorf("Fails to read param, error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	app, err := api.Scheduler.InspectApp(request.PathParameter("app_id"))
	if err != nil {
		logrus.Errorf("Get app error: %s", err.Error())
		response.WriteError(http.StatusNotFound, err)
		return
	}
	for index, weight := range param.Weights {
		indexInt, err := strconv.Atoi(index)
		if err != nil {
			logrus.Errorf("fails to update weight", err.Error())
			response.WriteError(http.StatusBadRequest, err)
			return
		}
		if slot, ok := app.Slots[indexInt]; ok {
			slot.SetWeight(weight)
		}
	}
	response.WriteEntity(FormAppRetWithVersions(app))

}

func (api *AppService) ProceedUpdate(request *restful.Request, response *restful.Response) {
	var param types.ProceedUpdateParam

	err := request.ReadEntity(&param)
	if err != nil {
		logrus.Errorf("Proceed update app error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	err = api.Scheduler.ProceedUpdate(request.PathParameter("app_id"), param.Instances, param.NewWeights)
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

func (api *AppService) UpdateAppTaskWeight(request *restful.Request, response *restful.Response) {
	var param types.UpdateWeightParam

	err := request.ReadEntity(&param)
	if err != nil {
		logrus.Errorf("update weight for a app error: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
	}

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

	slots := app.GetSlots()
	if task_index > len(slots)-1 || task_index < 0 {
		logrus.Errorf("slot not found: %d", task_index)
		response.WriteErrorString(http.StatusBadRequest, "Get task err: "+err.Error())
	}

	slot := slots[task_index]
	slot.SetWeight(param.Weight)

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

func (api *AppService) GetAppServiceDiscoveries(request *restful.Request, response *restful.Response) {
	app, err := api.Scheduler.InspectApp(request.PathParameter("app_id"))
	if err != nil {
		logrus.Errorf("Get app versions error: %s", err.Error())
		response.WriteError(http.StatusNotFound, err)
		return
	}
	serviceDiscoveries := getServiceDiscoveries(app)
	response.WriteEntity(serviceDiscoveries)
}

func (api *AppService) GetAppServiceDiscoveriesMD5(request *restful.Request, response *restful.Response) {
	app, err := api.Scheduler.InspectApp(request.PathParameter("app_id"))
	if err != nil {
		logrus.Errorf("Get app versions error: %s", err.Error())
		response.WriteError(http.StatusNotFound, err)
		return
	}
	serviceDiscoveries := getServiceDiscoveries(app)
	sd, err := json.Marshal(serviceDiscoveries)
	if err != nil {
		logrus.Errorf("Fails to Marshal serviceDiscoveries: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	sdMD5 := md5.Sum(sd)
	response.WriteEntity(hex.EncodeToString(sdMD5[:]))
}

func (api *AppService) GetAllServiceDiscoveries(request *restful.Request, response *restful.Response) {
	appFilterOptions := types.AppFilterOptions{}
	allServiceDiscoveries := make([]types.ServiceDiscovery, 0)
	for _, app := range api.Scheduler.ListApps(appFilterOptions) {
		serviceDiscoveries := getServiceDiscoveries(app)
		allServiceDiscoveries = append(allServiceDiscoveries, serviceDiscoveries...)
	}
	response.WriteEntity(allServiceDiscoveries)
}

func (api *AppService) GetAllServiceDiscoveriesMD5(request *restful.Request, response *restful.Response) {
	appFilterOptions := types.AppFilterOptions{}
	allServiceDiscoveries := make([]types.ServiceDiscovery, 0)
	for _, app := range api.Scheduler.ListApps(appFilterOptions) {
		serviceDiscoveries := getServiceDiscoveries(app)
		allServiceDiscoveries = append(allServiceDiscoveries, serviceDiscoveries...)
	}
	sd, err := json.Marshal(allServiceDiscoveries)
	if err != nil {
		logrus.Errorf("Fails to Marshal serviceDiscoveries: %s", err.Error())
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	sdMD5 := md5.Sum(sd)
	response.WriteEntity(hex.EncodeToString(sdMD5[:]))
}

func getServiceDiscoveries(app *state.App) []types.ServiceDiscovery {
	slots := app.GetSlots()
	serviceDiscoveries := make([]types.ServiceDiscovery, 0)
	if app.Mode == state.APP_MODE_FIXED {
		for _, slot := range slots {
			if slot.State == state.SLOT_STATE_TASK_RUNNING && slot.Healthy() {
				serviceDiscovery := types.ServiceDiscovery{
					TaskID:  slot.ID,
					AppID:   slot.App.ID,
					AppMode: string(slot.App.Mode),
					IP:      slot.Ip,
				}
				serviceDiscoveries = append(serviceDiscoveries, serviceDiscovery)
			}
		}
	} else {
		for _, slot := range slots {
			if slot.State == state.SLOT_STATE_TASK_RUNNING && slot.Healthy() {
				serviceDiscovery := types.ServiceDiscovery{
					TaskID:  slot.ID,
					AppID:   slot.App.ID,
					AppMode: string(slot.App.Mode),
					IP:      slot.AgentHostName,
					URL:     slot.ServiceDiscoveryURL(),
				}
				var taskPortMappings []*types.TaskPortMapping
				for i, hostPort := range slot.CurrentTask.HostPorts {
					taskPortMapping := &types.TaskPortMapping{
						HostPort: int32(hostPort),
					}
					appPortMapping := slot.App.CurrentVersion.Container.Docker.PortMappings[i]
					if appPortMapping != nil {
						taskPortMapping.ContainerPort = appPortMapping.ContainerPort
						taskPortMapping.Name = appPortMapping.Name
						taskPortMapping.Protocol = appPortMapping.Protocol
					}
					taskPortMappings = append(taskPortMappings, taskPortMapping)
				}
				serviceDiscovery.TaskPortMappings = taskPortMappings
				serviceDiscoveries = append(serviceDiscoveries, serviceDiscovery)
			}
		}
	}

	return serviceDiscoveries
}

func FormApp(app *state.App) *types.App {
	runningInstances := 0
	for _, slot := range app.GetSlots() {
		if slot.State == state.SLOT_STATE_TASK_RUNNING {
			runningInstances = runningInstances + 1
		}
	}
	version := app.CurrentVersion
	appRet := &types.App{
		ID:               app.ID,
		Name:             app.Name,
		Instances:        int(version.Instances),
		RunningInstances: runningInstances,
		RunAs:            version.RunAs,
		Priority:         int(version.Priority),
		ClusterID:        app.ClusterID,
		Created:          app.Created,
		Updated:          app.Updated,
		Mode:             string(app.Mode),
		State:            app.StateMachine.ReadableState(),
		CurrentVersion:   app.CurrentVersion,
		ProposedVersion:  app.ProposedVersion,
		Labels:           version.Labels,
		Env:              version.Env,
		Constraints:      version.Constraints,
		URIs:             version.URIs,
	}

	return appRet
}

func FormAppWithTask(app *state.App) *types.App {
	appRet := FormApp(app)

	// Add tasks without history info, should used by list API
	appRet.Tasks = FilterTasksFromApp(app)

	return appRet
}

func FormAppWithTaskHistory(app *state.App) *types.App {
	appRet := FormApp(app)

	// Add tasks with history info, should used by retrieve API
	appRet.Tasks = FilterTasksWithHistoryFromApp(app)

	return appRet
}

func FormAppRetWithVersions(app *state.App) *types.App {
	appRet := FormAppWithTaskHistory(app)
	appRet.Versions = make([]string, 0)
	for _, v := range app.Versions {
		appRet.Versions = append(appRet.Versions, v.ID)
	}
	return appRet
}

func FilterTasksFromApp(app *state.App) []*types.Task {
	tasks := make([]*types.Task, 0)
	for _, slot := range app.GetSlots() {
		task := FormTask(slot)
		tasks = append(tasks, task)
	}

	return tasks
}

func FilterTasksWithHistoryFromApp(app *state.App) []*types.Task {
	tasks := make([]*types.Task, 0)
	for _, slot := range app.GetSlots() {
		task := FormTask(slot)

		if len(slot.TaskHistory) > 0 {
			for _, v := range slot.TaskHistory {
				staleTask := &types.TaskHistory{
					ID:            v.ID,
					AppID:         app.ID,
					State:         v.State,
					Reason:        v.Reason,
					Message:       v.Message,
					OfferID:       v.OfferID,
					AgentID:       v.AgentID,
					AgentHostname: v.AgentHostName,

					Stderr: v.Stderr,
					Stdout: v.Stdout,

					ArchivedAt: v.ArchivedAt,
				}
				if v.Version != nil {
					staleTask.VersionID = v.Version.ID
					staleTask.AppVersion = v.Version.AppVersion
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

	task := FormTask(slot)

	if len(slot.TaskHistory) > 0 {
		for _, v := range slot.TaskHistory {
			if v == nil {
				continue
			}

			task.History = append(task.History, FormTaskHistory(v))
		}
	}

	return task, nil
}

func FormTaskHistory(v *state.Task) *types.TaskHistory {
	return &types.TaskHistory{
		ID:            v.ID,
		State:         v.State,
		Reason:        v.Reason,
		Message:       v.Message,
		OfferID:       v.OfferID,
		AgentID:       v.AgentID,
		AgentHostname: v.AgentHostName,
		VersionID:     v.Version.ID,
		ContainerId:   v.ContainerId,
		ContainerName: v.ContainerName,

		CPU:  v.Version.CPUs,
		Mem:  v.Version.Mem,
		Disk: v.Version.Disk,

		Stderr: v.Stderr,
		Stdout: v.Stdout,

		ArchivedAt: v.ArchivedAt,
	}
}

func FormTask(slot *state.Slot) *types.Task {
	task := &types.Task{
		ID:            slot.CurrentTask.ID,
		AppID:         slot.App.ID,
		SlotID:        slot.ID,
		VersionID:     slot.Version.ID,
		AppVersion:    slot.Version.AppVersion,
		Healthy:       slot.Healthy(),
		Status:        string(slot.State),
		OfferID:       slot.OfferID,
		AgentID:       slot.AgentID,
		AgentHostname: slot.AgentHostName,
		History:       make([]*types.TaskHistory, 0),
		CPU:           slot.Version.CPUs,
		Mem:           slot.Version.Mem,
		Disk:          slot.Version.Disk,
		IP:            slot.Ip,
		Ports:         slot.CurrentTask.HostPorts,
		Created:       slot.CurrentTask.Created,
		Image:         slot.Version.Container.Docker.Image,
		ContainerId:   slot.CurrentTask.ContainerId,
		ContainerName: slot.CurrentTask.ContainerName,
		Weight:        slot.GetWeight(),
	}
	return task
}
