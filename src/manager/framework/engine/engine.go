package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/src/util"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

type Engine struct {
	scontext         *swancontext.SwanContext
	heartbeater      *time.Ticker
	userEventChan    chan *event.UserEvent
	MesosCallChan    chan *sched.Call
	mesosFailureChan chan error

	handlerManager *HandlerManager

	stopC chan struct{}

	appLock sync.Mutex
	Apps    map[string]*state.App

	Allocator *state.OfferAllocator
	Scheduler *scheduler.Scheduler
}

func NewEngine(config util.SwanConfig) *Engine {
	engine := &Engine{
		Scheduler:     scheduler.NewScheduler(config.Scheduler),
		heartbeater:   time.NewTicker(10 * time.Second),
		userEventChan: make(chan *event.UserEvent, 1024), // TODO
		MesosCallChan: make(chan *sched.Call, 1024),      // 1024 TODO

		appLock: sync.Mutex{},
		Apps:    make(map[string]*state.App),
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

	engine.handlerManager = NewHanlderManager(engine, RegiserFun)
	engine.Allocator = state.NewOfferAllocator()
	return engine
}

// shutdown main engine and related
func (engine *Engine) Stop() error {
	engine.stopC <- struct{}{}
	return nil
}

// revive from crash or rotate from leader change
func (engine *Engine) Start() error {
	engine.Run(context.Background()) // context as a placeholder
	return nil
}

// main loop
func (engine *Engine) Run(ctx context.Context) error {
	if err := engine.Scheduler.ConnectToMesosAndAcceptEvent(); err != nil {
		logrus.Errorf("ConnectToMesosAndAcceptEvent got error %s", err)
		return err
	}

	for {
		select {
		case e := <-engine.Scheduler.MesosEventChan:
			logrus.WithFields(logrus.Fields{"mesos event chan": "yes"}).Debugf("")
			engine.handlerMesosEvent(e)
		case e := <-engine.userEventChan:
			fmt.Println(e)
			logrus.WithFields(logrus.Fields{"user event": "yes"}).Debugf("")
		case call := <-engine.MesosCallChan:
			engine.handlerMesosCall(call)
		case e := <-engine.mesosFailureChan:
			logrus.WithFields(logrus.Fields{"failure": "yes"}).Debugf("%s", e)
		case <-engine.heartbeater.C: // heartbeat timeout for now

		case <-engine.stopC:
			logrus.Infof("stopping main engine")
			return nil
		}
	}
}

func (engine *Engine) handlerMesosEvent(event *event.MesosEvent) {
	engine.handlerManager.Handle(event)
}

func (engine *Engine) handlerMesosCall(call *sched.Call) {
	logrus.WithFields(logrus.Fields{"sending-call": sched.Call_Type_name[int32(*call.Type)]}).Debugf("")
	resp, err := engine.Scheduler.Send(call)
	if err != nil {
		logrus.Errorf("%s", err)
	}
	if resp.StatusCode != 202 {
		logrus.Infof("send response not 202 but %d", resp.StatusCode)
	}
}
