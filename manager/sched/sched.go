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
		scheduler: scheduler.NewScheduler(config, scontext.Store),
	}

	backend := backend.NewBackend(s.scheduler, s.scontext.Store)
	s.scontext.ApiServer.AppendRouter(api.NewRouter(backend))
	return s
}

func (s *Sched) Start() error {
	if err := s.scheduler.Start(); err != nil {
		return err
	}

	return nil
}
