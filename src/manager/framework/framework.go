package framework

import (
	"sync"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/framework/api"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/framework/store"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"

	"golang.org/x/net/context"
)

type Framework struct {
	Scheduler   *scheduler.Scheduler
	SwanContext *swancontext.SwanContext
	RestApi     *api.AppService

	StopC chan struct{}
}

func New(SwanContext *swancontext.SwanContext, config config.SwanConfig, store store.Store, apiServer *apiserver.ApiServer) (*Framework, error) {
	f := &Framework{
		StopC:       make(chan struct{}),
		SwanContext: SwanContext,
	}

	f.Scheduler = scheduler.NewScheduler(config, SwanContext, store)
	f.RestApi = api.NewAndInstallAppService(apiServer, f.Scheduler)
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
