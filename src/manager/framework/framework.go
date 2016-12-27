package framework

import (
	"sync"

	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/framework/api"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/framework/store"

	"golang.org/x/net/context"
)

type Framework struct {
	Scheduler *scheduler.Scheduler
	RestApi   *api.AppService
	StatsApi  *api.StatsService
	EventsApi *api.EventsService

	StopC chan struct{}
}

func New(store store.Store, apiServer *apiserver.ApiServer) (*Framework, error) {
	f := &Framework{
		StopC: make(chan struct{}),
	}

	f.Scheduler = scheduler.NewScheduler(store)
	f.RestApi = api.NewAndInstallAppService(apiServer, f.Scheduler)
	f.StatsApi = api.NewAndInstallStatsService(apiServer, f.Scheduler)
	f.EventsApi = api.NewAndInstallEventsService(apiServer, f.Scheduler)
	return f, nil
}

func (f *Framework) Start(ctx context.Context) error {
	apiStopC := make(chan struct{})
	schedulerStopC := make(chan struct{})

	go func() {
		<-ctx.Done()
		apiStopC <- struct{}{}
		schedulerStopC <- struct{}{}
		f.StopC <- struct{}{}
	}()

	var err error
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		err = f.Scheduler.Start(ctx)
	}()

	wg.Wait()
	if err != nil {
		return err
	}

	return nil
}
