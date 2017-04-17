package framework

import (
	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/framework/api"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"

	"golang.org/x/net/context"
)

type Framework struct {
	Scheduler *scheduler.Scheduler
}

func New(apiServer *apiserver.ApiServer) (*Framework, error) {
	f := &Framework{}

	f.Scheduler = scheduler.NewScheduler()
	api.NewAndInstallAppService(apiServer, f.Scheduler)
	api.NewAndInstallStatsService(apiServer, f.Scheduler)
	api.NewAndInstallEventsService(apiServer, f.Scheduler)
	api.NewAndInstallHealthyService(apiServer, f.Scheduler)
	api.NewAndInstallFrameworkService(apiServer, f.Scheduler)
	api.NewAndInstallVersionService(apiServer, f.Scheduler)

	return f, nil
}

func (f *Framework) Start(ctx context.Context) error {
	errChan := make(chan error)
	go func() { errChan <- f.Scheduler.Start(ctx) }()
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
	}
	return nil
}

func (f *Framework) Stop() {
	f.Scheduler.Stop()
}
