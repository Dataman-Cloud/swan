package framework

import (
	"github.com/Dataman-Cloud/swan/src/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/framework/api"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/framework/store"
	"github.com/Dataman-Cloud/swan/src/swancontext"

	"golang.org/x/net/context"
)

type Framework struct {
	Scheduler *scheduler.Scheduler
	RestApi   *api.AppService
	StatsApi  *api.StatsService
	EventsApi *api.EventsService
	HealthApi *api.HealthyService

	StopC chan struct{}
}

func New(store store.Store, apiServer *apiserver.ApiServer) (*Framework, error) {
	f := &Framework{
		StopC: make(chan struct{}),
	}

	apiPrefix := swancontext.Instance().Config.ApiPrefix

	f.Scheduler = scheduler.NewScheduler(store)
	f.RestApi = api.NewAndInstallAppService(apiServer, f.Scheduler, apiPrefix)
	f.StatsApi = api.NewAndInstallStatsService(apiServer, f.Scheduler, apiPrefix)
	f.EventsApi = api.NewAndInstallEventsService(apiServer, f.Scheduler, apiPrefix)
	f.HealthApi = api.NewAndInstallHealthyService(apiServer, f.Scheduler, apiPrefix)
	return f, nil
}

func (f *Framework) Start(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() { errChan <- f.Scheduler.Start(ctx) }()

	for {
		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			f.StopC <- struct{}{}
		}
	}
}

func (f *Framework) Stop() {
	f.Scheduler.Stop()
}
