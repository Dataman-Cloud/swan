package scheduler

import ()

func SubscribedHandler(h *Handler) (*Handler, error) {
	sub := h.MesosEvent.Event.GetSubscribed()
	h.SchedulerRef.MesosConnector.Framework.Id = sub.FrameworkId

	return h, nil
}
