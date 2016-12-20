package scheduler

import (
	"golang.org/x/net/context"
)

func SubscribedHandler(h *Handler) (*Handler, error) {
	sub := h.MesosEvent.Event.GetSubscribed()
	h.Manager.SchedulerRef.MesosConnector.Framework.Id = sub.FrameworkId

	if err := h.Manager.SchedulerRef.store.UpdateFrameworkId(context.TODO(), *sub.FrameworkId.Value, nil); err != nil {
		return nil, err
	}

	return h, nil
}
