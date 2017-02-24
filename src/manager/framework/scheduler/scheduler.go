package scheduler

import (
	"time"

	swanevent "github.com/Dataman-Cloud/swan/src/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework/connector"
	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/manager/framework/store"
	"github.com/Dataman-Cloud/swan/src/swancontext"
	"github.com/Dataman-Cloud/swan/src/utils"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	CONNECTOR_DEFAULT_BACKOFF = 2 * time.Second
)

type Scheduler struct {
	heartbeater      *time.Ticker
	mesosFailureChan chan error
	stopC            chan struct{}

	handlerManager          *HandlerManager
	mesosConnectorCancelFun context.CancelFunc
	store                   store.Store

	AppStorage     *memoryStore
	UserEventChan  chan *event.UserEvent
	MesosConnector *connector.Connector
}

func NewScheduler(store store.Store) *Scheduler {
	scheduler := &Scheduler{
		MesosConnector: connector.NewConnector(),
		heartbeater:    time.NewTicker(10 * time.Second),

		AppStorage: NewMemoryStore(),
		store:      store,

		mesosFailureChan: make(chan error, 1),
		UserEventChan:    make(chan *event.UserEvent, 1024),
	}

	RegisterHandler := func(m *HandlerManager) {
		m.Register(event.EVENT_TYPE_MESOS_SUBSCRIBED, LoggerHandler, SubscribedHandler)
		m.Register(event.EVENT_TYPE_MESOS_HEARTBEAT, LoggerHandler, DummyHandler)
		m.Register(event.EVENT_TYPE_MESOS_OFFERS, LoggerHandler, OfferHandler, DummyHandler)
		m.Register(event.EVENT_TYPE_MESOS_RESCIND, LoggerHandler, DummyHandler)
		m.Register(event.EVENT_TYPE_MESOS_UPDATE, LoggerHandler, UpdateHandler, DummyHandler)
		m.Register(event.EVENT_TYPE_MESOS_FAILURE, LoggerHandler, DummyHandler)
		m.Register(event.EVENT_TYPE_MESOS_MESSAGE, LoggerHandler, DummyHandler)
		m.Register(event.EVENT_TYPE_MESOS_ERROR, LoggerHandler, DummyHandler)
		m.Register(event.EVENT_TYPE_USER_INVALID_APPS, LoggerHandler, InvalidAppHandler)
	}

	scheduler.handlerManager = NewHandlerManager(scheduler, RegisterHandler)

	state.SetStore(store)

	return scheduler
}

// shutdown main scheduler and related
func (scheduler *Scheduler) Stop() {
	scheduler.stopC <- struct{}{}
}

// revive from crash or rotate from leader change
func (scheduler *Scheduler) Start(ctx context.Context) error {
	if !swancontext.Instance().Config.NoRecover {
		apps, err := state.LoadAppData(scheduler.UserEventChan)
		if err != nil {
			return err
		}

		for _, app := range apps {
			scheduler.AppStorage.Add(app.ID, app)

			for _, slot := range app.GetSlots() {
				if slot.StateIs(state.SLOT_STATE_PENDING_OFFER) {
					state.OfferAllocatorInstance().PutSlotBackToPendingQueue(slot) // push the slot into pending offer queue
				}
			}
		}

		res, err := state.LoadOfferAllocatorMap()
		if err != nil {
			return err
		}

		state.OfferAllocatorInstance().AllocatedOffer = res

	}

	go func() {
		frameworkId, err := scheduler.store.GetFrameworkId()
		if err == nil {
			scheduler.MesosConnector.SetFrameworkInfoId(frameworkId)
		}

		var c context.Context
		c, scheduler.mesosConnectorCancelFun = context.WithCancel(ctx)
		scheduler.MesosConnector.Start(c, scheduler.mesosFailureChan)
	}()

	return scheduler.Run(context.Background()) // context as a placeholder
}

// main loop
func (scheduler *Scheduler) Run(ctx context.Context) error {
	for {
		select {
		case e := <-scheduler.MesosConnector.MesosEventChan:
			logrus.WithFields(logrus.Fields{"event": "mesos"}).Debugf("%s", e)
			scheduler.handleEvent(e)

		case e := <-scheduler.UserEventChan:
			logrus.WithFields(logrus.Fields{"event": "user"}).Debugf("%s", e)
			scheduler.handleEvent(e)

		case e := <-scheduler.mesosFailureChan:
			logrus.WithFields(logrus.Fields{"event": "mesosFailure"}).Debugf("%s", e)
			swanErr, ok := e.(*utils.SwanError)
			if ok && swanErr.Severity == utils.SeverityLow {
				time.Sleep(CONNECTOR_DEFAULT_BACKOFF)
				scheduler.MesosConnector.Reregister(scheduler.mesosFailureChan)
			} else {
				scheduler.mesosConnectorCancelFun()
				return e
			}

		case <-scheduler.heartbeater.C: // heartbeat timeout for now
			logrus.WithFields(logrus.Fields{"event": "heartBeat"}).Debugf("")

		case <-scheduler.stopC:
			logrus.WithFields(logrus.Fields{"event": "stopC"}).Debugf("")
			return nil
		}
	}
}

func (scheduler *Scheduler) handleEvent(e event.Event) {
	scheduler.handlerManager.Handle(e)
}

// reevaluation of apps state, clean up stale apps
func (scheduler *Scheduler) InvalidateApps() {
	appsPendingRemove := make([]string, 0)
	for _, app := range scheduler.AppStorage.Data() {
		if app.CanBeCleanAfterDeletion() { // check if app should be cleanup
			app.Remove()
			appsPendingRemove = append(appsPendingRemove, app.ID)
		}
	}

	for _, appId := range appsPendingRemove {
		scheduler.AppStorage.Delete(appId)
	}
}

func (scheduler *Scheduler) EmitEvent(swanEvent *swanevent.Event) {
	swancontext.Instance().EventBus.EventChan <- swanEvent
}

func (scheduler *Scheduler) HealthyTaskEvents() []*swanevent.Event {
	var healthyEvents []*swanevent.Event

	for _, app := range scheduler.AppStorage.Data() {
		for _, slot := range app.GetSlots() {
			if slot.Healthy() {
				healthyEvents = append(healthyEvents, slot.BuildTaskEvent(swanevent.EventTypeTaskHealthy))
			}
		}
	}

	return healthyEvents
}
