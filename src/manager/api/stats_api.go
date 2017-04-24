package api

import (
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/manager/connector"
	"github.com/Dataman-Cloud/swan/src/manager/scheduler"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/andygrunwald/megos"
	"github.com/emicklei/go-restful"

	"fmt"
	"net/url"
	"strings"
)

type StatsService struct {
	Scheduler *scheduler.Scheduler
}

func NewAndInstallStatsService(apiServer *apiserver.ApiServer, eng *scheduler.Scheduler) {
	statsService := &StatsService{
		Scheduler: eng,
	}
	apiserver.Install(apiServer, statsService)
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

	stats.ClusterID = connector.Instance().ClusterID

	appFilterOptions := types.AppFilterOptions{}
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

	master := strings.Split(connector.Instance().MesosLeader, "@")[1]
	node, _ := url.Parse(fmt.Sprintf("http://%s", master))
	state, _ := megos.NewClient([]*url.URL{node}, nil).GetStateFromCluster()

	slaves := make([]string, 0)
	for _, slave := range state.Slaves {
		stats.TotalCpu += slave.Resources.CPUs
		stats.TotalMem += slave.Resources.Mem
		stats.TotalDisk += slave.Resources.Disk

		s := strings.Split(slave.PID, "@")[1]
		slaves = append(slaves, s)

		if len(slave.Attributes) != 0 {
			stats.Attributes = append(stats.Attributes, slave.Attributes)
		}
	}

	stats.Created = state.StartTime
	stats.Master = strings.Split(state.Leader, "@")[1]
	stats.Slaves = strings.Join(slaves, " ")

	response.WriteEntity(stats)
}
