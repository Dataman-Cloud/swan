package scheduler

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
}

func NewHandler(id string, manager *HandlerManager, e *event.MesosEvent) *Handler {
	s := &Handler{
		Id:         id,
		MesosEvent: e,

		Manager: manager,
	}

	s.Response = NewResponse()
	return s
}

func (h *Handler) Process(timeoutCtx context.Context) {
	// remove this handler handlerManager
	defer func() {
		h.Manager.RemoveHandler(h.Id)
	}()

	select {
	case <-timeoutCtx.Done(): // abort early
		logrus.Errorf("%s", timeoutCtx.Err())
		return

	default:
		funcs := h.Manager.HandlerFuncs(h.MesosEvent.EventType)
		for _, fun := range funcs {
			h, err := fun(h)
			if err != nil {
				logrus.Errorf("%s, %s", h, err)
			}
		}

		for _, c := range h.Response.Calls {
			h.Manager.SchedulerRef.MesosConnector.MesosCallChan <- c
		}

		return
	}

}
