package scheduler

import (
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

func SubscribedHandler(h *Handler) (*Handler, error) {
	e, ok := h.Event.GetEvent().(*sched.Event)
	if !ok {
		logrus.Errorf("event conversion error %+v", h.Event)
		return h, nil
	}

	sub := e.GetSubscribed()
	h.Manager.SchedulerRef.MesosConnector.Framework.Id = sub.FrameworkId

	if err := h.Manager.SchedulerRef.store.UpdateFrameworkId(context.TODO(), *sub.FrameworkId.Value, nil); err != nil {
		return nil, err
	}

	return h, nil
}
