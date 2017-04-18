package scheduler

import (
	"errors"
	"reflect"
	"runtime"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Sirupsen/logrus"

	"golang.org/x/net/context"
)

var (
	errUnexpectedEventType = errors.New("unexpected event type")
)

type HandlerFunc func(s *Scheduler, e event.Event) error

func (hf HandlerFunc) Name() string {
	v := reflect.ValueOf(hf)
	return runtime.FuncForPC(v.Pointer()).Name()
}

type HandlerManager struct {
	handlerMap   map[string][]HandlerFunc // etype -> handlers...
	SchedulerRef *Scheduler               // FIX THIS Later
}

func NewHandlerManager(SchedulerRef *Scheduler) *HandlerManager {
	m := &HandlerManager{
		handlerMap:   make(map[string][]HandlerFunc),
		SchedulerRef: SchedulerRef,
	}

	m.Register(event.EVENT_TYPE_MESOS_SUBSCRIBED, LoggerHandler, SubscribedHandler)
	m.Register(event.EVENT_TYPE_MESOS_HEARTBEAT, LoggerHandler, DummyHandler)
	m.Register(event.EVENT_TYPE_MESOS_OFFERS, LoggerHandler, OfferHandler, DummyHandler)
	m.Register(event.EVENT_TYPE_MESOS_RESCIND, LoggerHandler, DummyHandler)
	m.Register(event.EVENT_TYPE_MESOS_UPDATE, LoggerHandler, UpdateHandler, DummyHandler)
	m.Register(event.EVENT_TYPE_MESOS_FAILURE, LoggerHandler, DummyHandler)
	m.Register(event.EVENT_TYPE_MESOS_MESSAGE, LoggerHandler, DummyHandler)
	m.Register(event.EVENT_TYPE_MESOS_ERROR, LoggerHandler, DummyHandler)
	m.Register(event.EVENT_TYPE_USER_INVALID_APPS, LoggerHandler, InvalidAppHandler)

	return m
}

func (m *HandlerManager) Register(eType string, funcs ...HandlerFunc) {
	m.handlerMap[eType] = funcs
}

func (m *HandlerManager) HandlerFuncs(eType string) []HandlerFunc {
	return m.handlerMap[eType]
}

func (m *HandlerManager) Handle(e event.Event) {
	timeoutCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	go m.Process(timeoutCtx, e)
}

func (m *HandlerManager) Process(timeoutCtx context.Context, e event.Event) {
	select {
	case <-timeoutCtx.Done(): // abort early
		logrus.Errorf("%s", timeoutCtx.Err())

	default:
		funcs := m.HandlerFuncs(e.GetEventType())
		for _, fun := range funcs {
			if err := fun(m.SchedulerRef, e); err != nil {
				logrus.Errorf("handler [%v] process event [%v] got error [%v]",
					HandlerFunc(fun).Name(), e, err)
			}
		}
	}
}
