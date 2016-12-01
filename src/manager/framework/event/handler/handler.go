package handler

import (
	"fmt"
	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
)

type Handler struct {
	Id         string
	Manager    *HandlerManager
	Response   *Response
	MesosEvent *event.MesosEvent
}

func NewHandler(id string, manager *HandlerManager, e *event.MesosEvent) *Handler {
	s := &Handler{
		Id:         id,
		Manager:    manager,
		MesosEvent: e,
	}

	s.Response = NewResponse()
	return s
}

func (h *Handler) Process() {
	defer func() {
		h.Manager.RemoveHandler(h.Id)
	}()

	funcs := h.Manager.HandlerFuncs(h.MesosEvent.EventType)
	for _, fun := range funcs {
		h = fun(h)
	}

	for _, c := range h.Response.Calls {
		fmt.Println(c)
	}

}
