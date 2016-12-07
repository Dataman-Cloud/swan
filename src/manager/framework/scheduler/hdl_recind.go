package scheduler

import (
	"github.com/Sirupsen/logrus"
)

func RecindHandler(h *Handler) (*Handler, error) {
	logrus.WithFields(logrus.Fields{"handler": "recind"}).Debugf("logger handler report got event type: %s", h.MesosEvent.EventType)
	return h, nil
}
