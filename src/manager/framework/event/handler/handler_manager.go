package handler

import (
	"sync"

	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
)

var once sync.Once

type HandlerFunc func(s *handler) *handler

type HandlerFuncs []HandlerFunc

type HandlerManager struct {
	handlers   map[string]*Handler
	handlerMap map[sched.Event_Type]HandlerFuncs
}

var manager *HandlerManager

func NewHanlderManager(installFun func(*HandlerManager)) *HandlerManager {
	once.Do(func() {
		manager := &HandlerManager{}
		installFun(manager)
	})

	return manager
}

func (m *HandlerManager) Register(etype sched.Event_Type, funcs ...HandlerFunc) {
	m.handlerMap[etype] = HandlerFuncs(funcs...)
}

func (m *HandlerManager) HandlerFuncs(etype sched.Event_Type) HandlerFuns {
	m.handlerMap[etype]
}

func (m *HandlerManager) Handle(e *event.MesosEvent) *Handler {
	return s
}

func (m *HandlerManager) RemoveHandler(handlerId string) {
}

func (m *HandlerManager) SweepHandlers() {
}
