package scheduler

import (
	"github.com/Sirupsen/logrus"
)

func InvalidAppHandler(h *Handler) (*Handler, error) {
	logrus.WithFields(logrus.Fields{"handler": "invalid offer"}).Debugf("logger handler report got event type: %s", h.Event.GetEventType())

	h.Manager.SchedulerRef.InvalidateApps()
	return h, nil
}
