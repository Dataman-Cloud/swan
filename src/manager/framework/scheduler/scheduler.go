package scheduler

import (
	"sync"
	"time"

	swanevent "github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework/mesos_connector"
	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/manager/framework/store"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/src/util"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

type Scheduler struct {
	scontext         *swancontext.SwanContext
	heartbeater      *time.Ticker
	mesosFailureChan chan error

	handlerManager *HandlerManager

	stopC chan struct{}

	appLock sync.Mutex
	Apps    map[string]*state.App

	Allocator      *state.OfferAllocator
	MesosConnector *mesos_connector.MesosConnector
	store          store.Store
}

func NewScheduler(config util.SwanConfig, scontext *swancontext.SwanContext, store store.Store) *Scheduler {
	scheduler := &Scheduler{
		MesosConnector: mesos_connector.NewMesosConnector(config.Scheduler),
		heartbeater:    time.NewTicker(10 * time.Second),
		scontext:       scontext,

		appLock: sync.Mutex{},
		Apps:    make(map[string]*state.App),
		store:   store,
	}

	RegiserFun := func(m *HandlerManager) {
		m.Register(sched.Event_SUBSCRIBED, LoggerHandler, SubscribedHandler)
		m.Register(sched.Event_HEARTBEAT, LoggerHandler, DummyHandler)
		m.Register(sched.Event_OFFERS, LoggerHandler, OfferHandler, DummyHandler)
		m.Register(sched.Event_RESCIND, LoggerHandler, DummyHandler)
		m.Register(sched.Event_UPDATE, LoggerHandler, UpdateHandler, DummyHandler)
		m.Register(sched.Event_FAILURE, LoggerHandler, DummyHandler)
		m.Register(sched.Event_MESSAGE, LoggerHandler, DummyHandler)
		m.Register(sched.Event_ERROR, LoggerHandler, DummyHandler)
	}

	scheduler.handlerManager = NewHanlderManager(scheduler, RegiserFun)
	scheduler.Allocator = state.NewOfferAllocator()

	return scheduler
}

// shutdown main scheduler and related
func (scheduler *Scheduler) Stop() error {
	scheduler.stopC <- struct{}{}
	return nil
}

// revive from crash or rotate from leader change
func (scheduler *Scheduler) Start(ctx context.Context) error {

	// temp solution
	go func() {
		scheduler.MesosConnector.Start(ctx)
	}()

	return scheduler.Run(context.Background()) // context as a placeholder
}

// main loop
func (scheduler *Scheduler) Run(ctx context.Context) error {
	if err := scheduler.MesosConnector.ConnectToMesosAndAcceptEvent(); err != nil {
		logrus.Errorf("ConnectToMesosAndAcceptEvent got error %s", err)
		return err
	}

	for {
		select {
		case e := <-scheduler.MesosConnector.MesosEventChan:
			logrus.WithFields(logrus.Fields{"mesos event chan": "yes"}).Debugf("")
			scheduler.handlerMesosEvent(e)

		case e := <-scheduler.mesosFailureChan:
			logrus.WithFields(logrus.Fields{"failure": "yes"}).Debugf("%s", e)

		case <-scheduler.heartbeater.C: // heartbeat timeout for now

		case <-scheduler.stopC:
			logrus.Infof("stopping main scheduler")
			return nil
		}
	}
}

func (scheduler *Scheduler) handlerMesosEvent(event *event.MesosEvent) {
	scheduler.handlerManager.Handle(event)
}

// reevaluation of apps state, clean up stale apps
func (scheduler *Scheduler) InvalidateApps() {
	appsPendingRemove := make([]string, 0)
	for _, app := range scheduler.Apps {
		if app.CanBeCleanAfterDeletion() { // check if app should be cleanup
			appsPendingRemove = append(appsPendingRemove, app.AppId)
		}
	}

	scheduler.appLock.Lock()
	defer scheduler.appLock.Unlock()
	for _, appId := range appsPendingRemove {
		delete(scheduler.Apps, appId)
	}
}

func (scheduler *Scheduler) EmitEvent(swanEvent *swanevent.Event) {
	scheduler.scontext.EventBus.EventChan <- swanEvent
}
