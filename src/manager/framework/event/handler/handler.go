package handler

import (
	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
)

type Handler struct {
	manager  *HandlerManager
	response *Response
}

func NewHandler(manager) *Handler {
	s := &Handler{}
	return s
}

func (s *Handler) Process() {
	funcs := s.manager.HandlerFuncs(s.event.Event_Type)
	for _, fun := range funcs {
		s = fun(s)
	}

	for _, c := range s.response.Calls {
	}
}
