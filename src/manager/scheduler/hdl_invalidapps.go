package scheduler

import (
	"github.com/Dataman-Cloud/swan/src/manager/event"

	"github.com/Sirupsen/logrus"
)

func InvalidAppHandler(s *Scheduler, ev event.Event) error {
	logrus.WithFields(logrus.Fields{"handler": "InvalidAppHandler"}).
		Debugf("logger handler report got event type: %s", ev.GetEventType())

	appID, ok := ev.GetEvent().(string)
	if ok {
		s.AppStorage.Delete(appID)
	}
	return nil
}
