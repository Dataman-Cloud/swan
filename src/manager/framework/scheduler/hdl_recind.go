package scheduler

import (
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"

	"github.com/Sirupsen/logrus"
)

func RecindHandler(h *Handler) (*Handler, error) {
	logrus.WithFields(logrus.Fields{"handler": "recind"}).Debugf("logger handler report got event type: %s", h.Event.GetEventType())

	_, ok := h.Event.GetEvent().(*sched.Event)
	if !ok {
		logrus.Errorf("event conversion error %+v", h.Event)
		return h, nil
	}

	return h, nil
}
