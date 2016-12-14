package scheduler

import (
	"sync"
	"time"

	"github.com/Dataman-Cloud/swan/src/config"
	swanevent "github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework/mesos_connector"
	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/manager/framework/store"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/src/types"

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
	config         config.SwanConfig
}

func NewScheduler(config config.SwanConfig, scontext *swancontext.SwanContext, store store.Store) *Scheduler {
	scheduler := &Scheduler{
		MesosConnector: mesos_connector.NewMesosConnector(config.Scheduler),
		heartbeater:    time.NewTicker(10 * time.Second),
		scontext:       scontext,

		appLock: sync.Mutex{},
		Apps:    make(map[string]*state.App),
		store:   store,
		config:  config,
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

	state.SetStore(store)

	return scheduler
}

// shutdown main scheduler and related
func (scheduler *Scheduler) Stop() error {
	scheduler.stopC <- struct{}{}
	return nil
}

// revive from crash or rotate from leader change
func (scheduler *Scheduler) Start(ctx context.Context) error {

	if !scheduler.config.NoRecover {
		if err := scheduler.LoadAppData(); err != nil {
			return err
		}
	}

	// temp solution
	go func() {
		scheduler.MesosConnector.Start(ctx)
	}()

	return scheduler.Run(context.Background()) // context as a placeholder
}

// load app data frm persistent data
func (scheduler *Scheduler) LoadAppData() error {
	raftApps, err := scheduler.store.ListApps()
	if err != nil {
		return err
	}

	apps := make(map[string]*state.App)

	for _, raftApp := range raftApps {
		app := &state.App{
			AppId:               raftApp.ID,
			CurrentVersion:      state.VersionFromRaft(raftApp.Version),
			State:               raftApp.State,
			Mode:                state.AppMode(raftApp.Version.Mode),
			Created:             time.Unix(0, raftApp.CreatedAt),
			Updated:             time.Unix(0, raftApp.UpdatedAt),
			Scontext:            scheduler.scontext,
			Slots:               make(map[int]*state.Slot),
			InvalidateCallbacks: make(map[string][]state.AppInvalidateCallbackFuncs),
			MesosConnector:      scheduler.MesosConnector,
			OfferAllocatorRef:   scheduler.Allocator,
		}

		raftVersions, err := scheduler.store.ListVersions(raftApp.ID)
		if err != nil {
			return err
		}

		var versions []*types.Version
		for _, raftVersion := range raftVersions {
			versions = append(versions, state.VersionFromRaft(raftVersion))
		}

		app.Versions = versions

		slots, err := scheduler.LoadAppSlots(app)
		if err != nil {
			return err
		}

		for _, slot := range slots {
			app.Slots[int(slot.Index)] = slot
		}

		apps[app.AppId] = app
	}

	scheduler.Apps = apps

	return nil
}

func (scheduler *Scheduler) LoadAppSlots(app *state.App) ([]*state.Slot, error) {
	raftSlots, err := scheduler.store.ListSlots(app.AppId)
	if err != nil {
		return nil, err
	}

	var slots []*state.Slot
	for _, raftSlot := range raftSlots {
		slot := state.SlotFromRaft(raftSlot)

		raftTasks, err := scheduler.store.ListTasks(app.AppId, slot.Id)
		if err != nil {
			return nil, err
		}

		var tasks []*state.Task
		for _, raftTask := range raftTasks {
			tasks = append(tasks, state.TaskFromRaft(raftTask))
		}
		slot.TaskHistory = tasks

		slot.CurrentTask.Slot = slot
		slot.CurrentTask.MesosConnector = app.MesosConnector

		if slot.CurrentTask.Version == nil {
			slot.CurrentTask.Version = app.CurrentVersion
		}
		slot.App = app

		//TODO: slot maybe not app currentVersion
		slot.Version = app.CurrentVersion

		slot.StatesCallbacks = make(map[string][]state.SlotStateCallbackFuncs)

		slots = append(slots, slot)
	}

	return slots, nil
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
