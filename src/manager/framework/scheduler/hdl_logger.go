package scheduler

import (
	"github.com/Sirupsen/logrus"
)

func LoggerHandler(h *Handler) (*Handler, error) {
	logrus.WithFields(logrus.Fields{"handler": "logger"}).Debugf("logger handler report got event type: %s", h.Event.GetEventType())
	return h, nil
}
