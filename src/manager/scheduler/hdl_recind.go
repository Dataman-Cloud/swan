package scheduler

import (
	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"

	"github.com/Sirupsen/logrus"
)

func RecindHandler(s *Scheduler, ev event.Event) error {
	logrus.WithFields(logrus.Fields{"handler": "recind"}).
		Debugf("logger handler report got event type: %s", ev.GetEventType())

	if _, ok := ev.GetEvent().(*sched.Event); !ok {
		return errUnexpectedEventType
	}

	return nil
}
