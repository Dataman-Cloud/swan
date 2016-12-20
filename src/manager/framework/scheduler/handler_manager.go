package scheduler

import (
	"sync"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"

	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

var once sync.Once

type HandlerFunc func(s *Handler) (*Handler, error)

type HandlerFuncs []HandlerFunc

type HandlerManager struct {
	lock       sync.Mutex
	handlers   map[string]*Handler
	handlerMap map[sched.Event_Type]HandlerFuncs

	SchedulerRef *Scheduler
}

func NewHanlderManager(scheduler *Scheduler, installFun func(*HandlerManager)) *HandlerManager {
	manager := &HandlerManager{
		handlers:     make(map[string]*Handler),
		handlerMap:   make(map[sched.Event_Type]HandlerFuncs),
		lock:         sync.Mutex{},
		SchedulerRef: scheduler,
	}
	once.Do(func() {
		installFun(manager)
	})

	return manager
}

func (m *HandlerManager) Register(etype sched.Event_Type, funcs ...HandlerFunc) {
	m.handlerMap[etype] = HandlerFuncs(funcs)
}

func (m *HandlerManager) HandlerFuncs(etype sched.Event_Type) HandlerFuncs {
	return m.handlerMap[etype]
}

func (m *HandlerManager) Handle(e *event.MesosEvent) *Handler {
	handlerId := uuid.NewV4().String()
	h := NewHandler(handlerId, m, e)

	m.lock.Lock()
	defer m.lock.Unlock()
	m.handlers[handlerId] = h

	timeoutCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	go h.Process(timeoutCtx) // process a mesos event in seperated goroutine

	return h
}

func (m *HandlerManager) RemoveHandler(handlerId string) {
	m.lock.Lock()
	defer m.lock.Unlock() // protect mutual access to m.handlers

	delete(m.handlers, handlerId)
}
