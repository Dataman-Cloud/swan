package engine

import (
	"github.com/Dataman-Cloud/swan/src/manager/framework/event"

	"github.com/Sirupsen/logrus"
)

type Handler struct {
	Id         string
	Manager    *HandlerManager
	Response   *Response
	MesosEvent *event.MesosEvent

	EngineRef *Engine
}

func NewHandler(id string, manager *HandlerManager, e *event.MesosEvent) *Handler {
	s := &Handler{
		Id:         id,
		Manager:    manager,
		MesosEvent: e,
		EngineRef:  manager.EngineRef,
	}

	s.Response = NewResponse()
	return s
}

func (h *Handler) Process() {
	// remove this handler
	defer func() {
		h.Manager.RemoveHandler(h.Id)
	}()

	funcs := h.Manager.HandlerFuncs(h.MesosEvent.EventType)
	for _, fun := range funcs {
		h, err := fun(h)
		if err != nil {
			logrus.Errorf("%s, %s", h, err)
		}
	}

	for _, c := range h.Response.Calls {
		h.EngineRef.MesosCallChan <- c
	}
}
