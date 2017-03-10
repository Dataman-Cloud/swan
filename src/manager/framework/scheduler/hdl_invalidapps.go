package scheduler

import (
	"github.com/Sirupsen/logrus"
)

func InvalidAppHandler(h *Handler) (*Handler, error) {
	logrus.WithFields(logrus.Fields{"handler": "InvalidAppHandler"}).Debugf("logger handler report got event type: %s", h.Event.GetEventType())

	appID, ok := h.Event.GetEvent().(string)

	if ok {
		h.Manager.SchedulerRef.AppStorage.Delete(appID)
	}
	return h, nil
}
