package scheduler

import (
	"github.com/Dataman-Cloud/swan/src/manager/connector"
	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/store"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"

	"github.com/Sirupsen/logrus"
)

func SubscribedHandler(s *Scheduler, ev event.Event) error {
	e, ok := ev.GetEvent().(*sched.Event)
	if !ok {
		return errUnexpectedEventType
	}

	logrus.Infof("subscribed successful with ID %s", e.GetSubscribed().FrameworkId.GetValue())

	sub := e.GetSubscribed()
	connector.Instance().SetFrameworkInfoId(*sub.FrameworkId.Value)

	return store.DB().UpdateFrameworkId(*sub.FrameworkId.Value)
}
