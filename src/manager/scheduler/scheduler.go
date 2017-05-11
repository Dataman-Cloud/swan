package scheduler

import (
	"time"

	"github.com/Dataman-Cloud/swan/src/config"
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

	handlerManager          *HandlerManager
	mesosConnectorCancelFun context.CancelFunc

	userEventChan chan *event.UserEvent

	AppStorage     *memoryStore
	MesosConnector *connector.Connector
}

func NewScheduler(cfg config.ManagerConfig) *Scheduler {
	connector.Init(cfg.MesosFrameworkUser, cfg.MesosURL)

	sched := &Scheduler{
		MesosConnector: connector.Instance(),
		heartbeater:    time.NewTicker(10 * time.Second),

		AppStorage: NewMemoryStore(),

		userEventChan: make(chan *event.UserEvent, 1024),
	}

	sched.handlerManager = NewHandlerManager(sched)

	return sched
}

// shutdown main scheduler and related
// revive from crash or rotate from leader change
func (sched *Scheduler) Start(ctx context.Context) error {
	if err := sched.recoverFromPreviousScene(); err != nil {
		return err
	}

	go func() {
		sched.MesosConnector.SetFrameworkInfoId(store.DB().GetFrameworkId())

		var c context.Context
		c, sched.mesosConnectorCancelFun = context.WithCancel(ctx)
		sched.MesosConnector.Start(c)
	}()

	for {
		select {
		case e := <-sched.userEventChan:
			logrus.WithFields(logrus.Fields{"event": "user"}).Debugf("%s", e)
			sched.handleEvent(e)

		case e := <-sched.MesosConnector.MesosEvent(): // subcribe connector's mesos-events
			logrus.WithFields(logrus.Fields{"event": "mesos"}).Debugf("%s", e)
			sched.handleEvent(e)

		// TODO: make the connector self-contains rejoin logic
		case e := <-sched.MesosConnector.ErrEvent(): // subcribe connector's failures events
			logrus.WithFields(logrus.Fields{"event": "mesosFailure"}).Errorf("%s", e)
			swanErr, ok := e.(*utils.SwanError)
			if ok && swanErr.Severity == utils.SeverityLow {
				for {
					time.Sleep(CONNECTOR_DEFAULT_BACKOFF)
					err := sched.MesosConnector.Reregister() // CAUTION
					if err == nil {
						break
					}
				}
			} else {
				sched.mesosConnectorCancelFun() // CAUTION
				return e
			}

		case <-sched.heartbeater.C: // heartbeat timeout for now
			logrus.WithFields(logrus.Fields{"event": "heartBeat"}).Debugln("heart beat package")

		case <-ctx.Done():
			logrus.Info("scheduler shutdown  goroutine by ctx cancel")
			return ctx.Err()
		}
	}
}

func (sched *Scheduler) recoverFromPreviousScene() error {
	err := store.DB().Recover()
	if err != nil {
		return err
	}

	apps := state.LoadAppData(sched.userEventChan)
	for _, app := range apps {
		sched.AppStorage.Add(app.ID, app)

		for _, slot := range app.GetSlots() {
			if slot.StateIs(state.SLOT_STATE_PENDING_OFFER) {
				state.OfferAllocatorInstance().PutSlotBackToPendingQueue(slot) // push the slot into pending offer queue
			}
		}
	}

	state.OfferAllocatorInstance().AllocatedOffer, err = state.LoadOfferAllocatorMap()
	if err != nil {
		return err
	}

	return nil
}

func (sched *Scheduler) handleEvent(e event.Event) {
	sched.handlerManager.Handle(e)
}

func (sched *Scheduler) HealthyTaskEvents() []*eventbus.Event {
	var healthyEvents []*eventbus.Event

	for _, app := range sched.AppStorage.Data() {
		for _, slot := range app.GetSlots() {
			if slot.Healthy() {
				healthyEvents = append(healthyEvents, slot.BuildTaskEvent(eventbus.EventTypeTaskHealthy))
			}
		}
	}

	return healthyEvents
}
