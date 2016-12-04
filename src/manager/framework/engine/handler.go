package engine

import (
	"github.com/Dataman-Cloud/swan/src/manager/framework/event"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
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

func (h *Handler) Process(timeoutCtx context.Context) {
	// remove this handler handlerManager
	defer func() {
		h.Manager.RemoveHandler(h.Id)
	}()

	for {
		select {
		case <-timeoutCtx.Done(): // abort early
			// check timeoutCtx.Err()
			// if timeout
			// else cancelled
		default:
			funcs := h.Manager.HandlerFuncs(h.MesosEvent.EventType)
			for _, fun := range funcs {
				h, err := fun(h)
				if err != nil {
					logrus.Errorf("%s, %s", h, err)
				}
			}

			for _, c := range h.Response.Calls {
				h.EngineRef.Scheduler.MesosCallChan <- c
			}

			return
		}
	}

}
