package engine

import ()

func SubscribedHandler(h *Handler) (*Handler, error) {
	sub := h.MesosEvent.Event.GetSubscribed()
	h.EngineRef.Scheduler.Framework.Id = sub.FrameworkId

	return h, nil
}
