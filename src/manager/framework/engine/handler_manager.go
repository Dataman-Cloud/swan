package engine

import (
	"sync"

	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"

	"github.com/satori/go.uuid"
)

var once sync.Once

type HandlerFunc func(s *Handler) (*Handler, error)

type HandlerFuncs []HandlerFunc

type HandlerManager struct {
	lock       sync.Mutex
	handlers   map[string]*Handler
	handlerMap map[sched.Event_Type]HandlerFuncs

	EngineRef *Engine
}

func NewHanlderManager(engine *Engine, installFun func(*HandlerManager)) *HandlerManager {
	manager := &HandlerManager{
		handlers:   make(map[string]*Handler),
		handlerMap: make(map[sched.Event_Type]HandlerFuncs),
		lock:       sync.Mutex{},
		EngineRef:  engine,
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
	m.handlers[handlerId] = h
	m.lock.Unlock()

	go h.Process()

	return h
}

func (m *HandlerManager) RemoveHandler(handlerId string) {
	delete(m.handlers, handlerId)
}
