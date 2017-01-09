package api

import (
	"github.com/Dataman-Cloud/swan/src/apiserver"
	"github.com/Dataman-Cloud/swan/src/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/framework/mesos_connector"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/emicklei/go-restful"
)

type StatsService struct {
	Scheduler *scheduler.Scheduler
	apiserver.ApiRegister
}

func NewAndInstallStatsService(apiServer *apiserver.ApiServer, eng *scheduler.Scheduler) *StatsService {
	statsService := &StatsService{
		Scheduler: eng,
	}
	apiserver.Install(apiServer, statsService)
	return statsService
}

func (api *StatsService) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		ApiVersion(config.API_PREFIX).
		Path("/stats").
		Doc("stats API").
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(metrics.InstrumentRouteFunc("GET", "Stats", api.Stats)).
		// docs
		Doc("Get Stats").
		Operation("getStats").
		Returns(200, "OK", types.Stats{}))

	container.Add(ws)
}

func (api *StatsService) Stats(request *restful.Request, response *restful.Response) {
	var stats types.Stats
	stats.AppStats = make(map[string]int)

	stats.ClusterID = mesos_connector.Instance().ClusterID

	appFilterOptions := scheduler.AppFilterOptions{}
	for _, app := range api.Scheduler.ListApps(appFilterOptions) {
		version := app.CurrentVersion
		stats.AppCount += 1
		stats.AppStats[version.RunAs] += 1

		stats.TaskCount += int(version.Instances)

		for _, slot := range app.GetSlots() {
			stats.CpuTotalOffered += slot.ResourcesUsed().CPU
			stats.MemTotalOffered += slot.ResourcesUsed().Mem
			stats.DiskTotalOffered += slot.ResourcesUsed().Disk

			// TODO(xychu): add usage stats
		}
	}

	response.WriteEntity(stats)
}
