package middleware

import (
	"github.com/Dataman-Cloud/swan/src/manager/framework/event/handler"

	"github.com/Sirupsen/logrus"
)

func DummyHandler(h *handler.Handler) *handler.Handler {
	logrus.WithFields(logrus.Fields{"handler": "dummy"}).Debugf("dummy handler report got event type: %s", h.MesosEvent.EventType)
	return h
}
