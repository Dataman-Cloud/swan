package sched

import (
	"github.com/Dataman-Cloud/swan/manager/sched/api"
	"github.com/Dataman-Cloud/swan/manager/sched/backend"
	"github.com/Dataman-Cloud/swan/manager/sched/scheduler"
	"github.com/Dataman-Cloud/swan/manager/swancontext"
	"github.com/Dataman-Cloud/swan/util"
)

type Sched struct {
	config    util.Scheduler
	scheduler *scheduler.Scheduler
	scontext  *swancontext.SwanContext
}

func New(config util.Scheduler, scontext *swancontext.SwanContext) *Sched {
	s := &Sched{config: config,
		scontext:  scontext,
		scheduler: scheduler.New(config, scontext),
	}

	backend := backend.NewBackend(s.scheduler, s.scontext.Store)
	s.scontext.ApiServer.AppendRouter(api.NewRouter(backend))
	return s
}

func (s *Sched) Start() {
}
func (s *Sched) Run() error {
	return s.scheduler.Run()
}
