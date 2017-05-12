package api

import (
	"fmt"
	"time"

	"github.com/Dataman-Cloud/swan/src/config"
	apiserver "github.com/Dataman-Cloud/swan/src/manager/api/server"
	"github.com/Dataman-Cloud/swan/src/manager/api/server/metrics"
	"github.com/Dataman-Cloud/swan/src/manager/compose"
	"github.com/Dataman-Cloud/swan/src/manager/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/store"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/emicklei/go-restful"
	uuid "github.com/satori/go.uuid"
)

type ComposeService struct {
	sched *scheduler.Scheduler
}

func NewAndInstallComposeService(server *apiserver.ApiServer, sched *scheduler.Scheduler) {
	apiserver.Install(server, &ComposeService{sched})
}

func (api *ComposeService) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.ApiVersion(config.API_PREFIX).
		Path(config.API_PREFIX + "/compose").
		Doc("compose API").
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/").Doc("Run Compose Instance").
		To(metrics.InstrumentRouteFunc("POST", "Compose", api.runInstance)))

	ws.Route(ws.POST("/parse").Doc("Parse Compose YAML Text").
		To(metrics.InstrumentRouteFunc("POST", "Parse", api.parseYAML)))

	ws.Route(ws.GET("/").Doc("List Compose Instances").
		To(metrics.InstrumentRouteFunc("GET", "Compose", api.listInstances)))

	ws.Route(ws.GET("/{iid}").Doc("Get Compose Instance").
		To(metrics.InstrumentRouteFunc("GET", "Compose", api.getInstance)).
		Param(ws.PathParameter("iid", "id or name of compose instance").DataType("string")))

	ws.Route(ws.DELETE("/{iid}").Doc("Delete Compose Instance").
		To(metrics.InstrumentRouteFunc("DELETE", "Compose", api.removeInstance)).
		Param(ws.PathParameter("iid", "id or name of comopse instance").DataType("string")))

	container.Add(ws)
}

func (api *ComposeService) listInstances(r *restful.Request, w *restful.Response) {
	is, err := store.DB().ListInstances()
	if err != nil {
		w.WriteError(500, err)
		return
	}
	w.WriteEntity(is)
}

func (api *ComposeService) getInstance(r *restful.Request, w *restful.Response) {
	id := r.PathParameter("iid") // id or name
	i, err := store.DB().GetInstance(id)
	if err != nil {
		w.WriteError(500, err)
		return
	}

	w.WriteEntity(api.newInstanceWrapper(i))
}

// parse yaml content and output service & variable definations
func (api *ComposeService) parseYAML(r *restful.Request, w *restful.Response) {
	var req struct {
		YAML string `json:"yaml"`
	}
	if err := r.ReadEntity(&req); err != nil {
		w.WriteError(400, err)
		return
	}

	cfg, err := compose.YamlServices([]byte(req.YAML), nil)
	if err != nil {
		w.WriteError(400, err)
		return
	}

	var (
		vars = compose.YamlVariables([]byte(req.YAML))
		svrs = make([]string, 0, 0)
	)
	for _, svr := range cfg.Services {
		svrs = append(svrs, svr.Name)
	}

	w.WriteHeaderAndEntity(201, map[string]interface{}{
		"services":  svrs,
		"variables": vars,
	})
}

// launch compose instance
func (api *ComposeService) runInstance(r *restful.Request, w *restful.Response) {
	var err error

	// obtain & verify
	var ins *store.Instance
	if err = r.ReadEntity(&ins); err != nil {
		w.WriteError(400, err)
		return
	}
	// raw docker compose only, convert & rewrite service groups
	if ins.RequireConvert() {
		ins.ServiceGroup, err = compose.YamlToServiceGroup(
			[]byte(ins.YAMLRaw), // docker-compose yaml raw
			ins.YAMLEnv,         // docker-compose yaml env
			ins.YAMLExtra,       // extra compose settings
		)
		if err != nil {
			w.WriteError(400, fmt.Errorf("yaml convert: %v", err))
			return
		}
	}

	// verify
	if err := ins.Valid(); err != nil {
		w.WriteError(400, err)
		return
	}

	// ensure all settings could be converted to types.Version to fit with state.NewApp()
	for name, svr := range ins.ServiceGroup {
		if _, err := compose.SvrToVersion(svr, "", ""); err != nil {
			w.WriteError(400, fmt.Errorf("convert svr %s error: %v", name, err))
			return
		}
	}

	// check conflict
	if ins, _ := store.DB().GetInstance(ins.Name); ins != nil {
		w.WriteHeader(409)
		return
	}

	// db save
	ins.ID = uuid.NewV4().String()
	ins.Status = "creating"
	ins.CreatedAt = time.Now()
	ins.UpdatedAt = time.Now()
	if err := store.DB().CreateInstance(ins); err != nil {
		w.WriteError(500, err)
		return
	}

	// async launch
	go compose.LaunchInstance(ins, api.sched)
	w.WriteHeaderAndEntity(201, ins)
}

// remove compose instance
func (api *ComposeService) removeInstance(r *restful.Request, w *restful.Response) {
	id := r.PathParameter("iid") // id or name
	i, err := store.DB().GetInstance(id)
	if err != nil {
		w.WriteError(500, err)
		return
	}

	// remove apps
	apps := compose.InstanceApps(api.sched, i.Name)
	for _, app := range apps {
		if err := api.sched.DeleteApp(app.ID); err != nil {
			w.WriteError(500, fmt.Errorf("remove instance:%s app:%s error: %v",
				i.Name, app.Name, err))
			return
		}
	}

	// remove instance
	if err := store.DB().DeleteInstance(i.ID); err != nil {
		w.WriteError(500, err)
		return
	}

	w.WriteHeader(204)
}

type instanceWrapper struct {
	*store.Instance
	Apps []*types.App `json:"apps"`
}

func (api *ComposeService) newInstanceWrapper(i *store.Instance) *instanceWrapper {
	var (
		// sigh... state.App can't be Marshal, as loop references cause OOM
		as   = compose.InstanceApps(api.sched, i.Name)
		apps = make([]*types.App, 0, len(as))
	)
	for _, a := range as {
		apps = append(apps, FormApp(a))
	}

	return &instanceWrapper{
		Instance: i,
		Apps:     apps,
	}
}
