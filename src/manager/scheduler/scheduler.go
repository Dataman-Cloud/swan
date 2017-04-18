package scheduler

import (
	"time"

	eventbus "github.com/Dataman-Cloud/swan/src/event"
	"github.com/Dataman-Cloud/swan/src/manager/connector"
	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/state"
	"github.com/Dataman-Cloud/swan/src/manager/store"
	"github.com/Dataman-Cloud/swan/src/utils"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	CONNECTOR_DEFAULT_BACKOFF = 2 * time.Second
)

type Scheduler struct {
	heartbeater *time.Ticker
	stopC       chan struct{}

	handlerManager          *HandlerManager
	mesosConnectorCancelFun context.CancelFunc
	store                   store.Store

	userEventChan chan *event.UserEvent

	AppStorage     *memoryStore
	MesosConnector *connector.Connector
}

func NewScheduler() *Scheduler {
	scheduler := &Scheduler{
		stopC:          make(chan struct{}),
		MesosConnector: connector.Instance(),
		heartbeater:    time.NewTicker(10 * time.Second),

		AppStorage: NewMemoryStore(),
		store:      store.DB(),

		userEventChan: make(chan *event.UserEvent, 1024),
	}

	scheduler.handlerManager = NewHandlerManager(scheduler)

	return scheduler
}

// shutdown main scheduler and related
func (scheduler *Scheduler) Stop() {
	close(scheduler.stopC)
}

// revive from crash or rotate from leader change
func (scheduler *Scheduler) Start(ctx context.Context) error {
	apps, err := state.LoadAppData(scheduler.userEventChan)
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

	go func() {
		//frameworkId, err := scheduler.store.GetFrameworkId()
		//if err == nil {
		//scheduler.MesosConnector.SetFrameworkInfoId(frameworkId)
		//}

		var c context.Context
		c, scheduler.mesosConnectorCancelFun = context.WithCancel(ctx)
		scheduler.MesosConnector.Start(c)
	}()

	return scheduler.Run(context.Background()) // context as a placeholder
}

// main loop
func (scheduler *Scheduler) Run(ctx context.Context) error {
	for {
		select {
		case e := <-scheduler.userEventChan:
			logrus.WithFields(logrus.Fields{"event": "user"}).Debugf("%s", e)
			scheduler.handleEvent(e)

		case e := <-scheduler.MesosConnector.MesosEvent(): // subcribe connector's mesos-events
			logrus.WithFields(logrus.Fields{"event": "mesos"}).Debugf("%s", e)
			scheduler.handleEvent(e)

		// TODO: make the connector self-contains rejoin logic
		case e := <-scheduler.MesosConnector.ErrEvent(): // subcribe connector's failures events
			logrus.WithFields(logrus.Fields{"event": "mesosFailure"}).Errorf("%s", e)
			swanErr, ok := e.(*utils.SwanError)
			if ok && swanErr.Severity == utils.SeverityLow {
				for {
					time.Sleep(CONNECTOR_DEFAULT_BACKOFF)
					err := scheduler.MesosConnector.Reregister() // CAUTION
					if err == nil {
						break
					}
				}
			} else {
				scheduler.mesosConnectorCancelFun() // CAUTION
				return e
			}

		case <-scheduler.heartbeater.C: // heartbeat timeout for now
			logrus.WithFields(logrus.Fields{"event": "heartBeat"}).Debugln("I'm alive")

		case <-scheduler.stopC:
			logrus.WithFields(logrus.Fields{"event": "stopC"}).Infoln("scheduler stopped")
			return nil
		}
	}
}

func (scheduler *Scheduler) handleEvent(e event.Event) {
	scheduler.handlerManager.Handle(e)
}

func (scheduler *Scheduler) HealthyTaskEvents() []*eventbus.Event {
	var healthyEvents []*eventbus.Event

	for _, app := range scheduler.AppStorage.Data() {
		for _, slot := range app.GetSlots() {
			if slot.Healthy() {
				healthyEvents = append(healthyEvents, slot.BuildTaskEvent(eventbus.EventTypeTaskHealthy))
			}
		}
	}

	return healthyEvents
}
