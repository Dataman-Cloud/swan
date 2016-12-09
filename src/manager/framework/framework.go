package framework

import (
	"sync"

	"github.com/Dataman-Cloud/swan/src/manager/framework/api"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/framework/store"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"

	"github.com/Dataman-Cloud/swan/src/util"

	"golang.org/x/net/context"
)

type Framework struct {
	Scheduler   *scheduler.Scheduler
	HttpApi     *api.Api
	SwanContext *swancontext.SwanContext

	StopC chan struct{}
}

func New(SwanContext *swancontext.SwanContext, config util.SwanConfig, store store.Store) (*Framework, error) {
	f := &Framework{
		StopC:       make(chan struct{}),
		SwanContext: SwanContext,
	}

	f.Scheduler = scheduler.NewScheduler(config, SwanContext, store)
	f.HttpApi = api.NewApi(f.Scheduler, config.Scheduler)

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
	wg.Add(2)
	go func() {
		wg.Done()
		err = f.HttpApi.Start(ctx)
	}()

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
