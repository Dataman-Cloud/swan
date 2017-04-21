package scheduler

import (
	"github.com/Dataman-Cloud/swan/src/manager/event"

	"github.com/Sirupsen/logrus"
)

func LoggerHandler(s *Scheduler, ev event.Event) error {
	switch typ := ev.GetEventType(); typ {

	case event.EVENT_TYPE_MESOS_ERROR:
		logrus.WithFields(logrus.Fields{"handler": "logger"}).
			Errorf("logger handler got event: %s", ev)

	case event.EVENT_TYPE_MESOS_FAILURE:
		logrus.WithFields(logrus.Fields{"handler": "logger"}).
			Warnf("logger handler got event: %s", ev)

	default:
		logrus.WithFields(logrus.Fields{"handler": "logger"}).
			Debugf("logger handler got event: %s", typ)
	}

	return nil
}
