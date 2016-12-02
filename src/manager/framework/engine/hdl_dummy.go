package engine

import (
	"github.com/Sirupsen/logrus"
)

func DummyHandler(h *Handler) *Handler {
	logrus.WithFields(logrus.Fields{"handler": "dummy"}).Debugf("dummy handler report got event type: %s", h.MesosEvent.EventType)
	return h
}
