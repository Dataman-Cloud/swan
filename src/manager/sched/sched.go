package sched

import (
	"github.com/Dataman-Cloud/swan/src/manager/sched/api"
	"github.com/Dataman-Cloud/swan/src/manager/sched/backend"
	"github.com/Dataman-Cloud/swan/src/manager/sched/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"
	"github.com/Dataman-Cloud/swan/src/util"

	"golang.org/x/net/context"
)

type Sched struct {
	config    util.Scheduler
	scheduler *scheduler.Scheduler
	scontext  *swancontext.SwanContext
}

func New(config util.Scheduler, scontext *swancontext.SwanContext) *Sched {
	s := &Sched{config: config,
		scontext:  scontext,
		scheduler: scheduler.NewScheduler(config, scontext.Store),
	}

	backend := backend.NewBackend(s.scheduler, s.scontext.Store)
	s.scontext.ApiServer.AppendRouter(api.NewRouter(backend))
	return s
}

//TODO (niumingguo) stop handle event when parent call cancel
func (s *Sched) Start(ctx context.Context) error {
	go func() {
		select {
		case <-ctx.Done():
			s.Stop()
		}
	}()

	return s.scheduler.Start()
}

func (s *Sched) Stop() {
	s.scheduler.Stop()
}
