package scheduler

import (
	"sync"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/framework/event"

	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

var once sync.Once

type HandlerFunc func(s *Handler) (*Handler, error)

type HandlerFuncs []HandlerFunc

type HandlerManager struct {
	lock       sync.Mutex
	handlers   map[string]*Handler
	handlerMap map[string]HandlerFuncs

	SchedulerRef *Scheduler
}

func NewHanlderManager(scheduler *Scheduler, installFun func(*HandlerManager)) *HandlerManager {
	manager := &HandlerManager{
		handlers:     make(map[string]*Handler),
		handlerMap:   make(map[string]HandlerFuncs),
		lock:         sync.Mutex{},
		SchedulerRef: scheduler,
	}
	once.Do(func() {
		installFun(manager)
	})

	return manager
}

func (m *HandlerManager) Register(etype string, funcs ...HandlerFunc) {
	m.handlerMap[etype] = HandlerFuncs(funcs)
}

func (m *HandlerManager) HandlerFuncs(etype string) HandlerFuncs {
	return m.handlerMap[etype]
}

func (m *HandlerManager) Handle(e event.Event) *Handler {
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
