package scheduler

import (
	"github.com/Dataman-Cloud/swan/src/manager/framework/connector"
	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

func SubscribedHandler(s *Scheduler, ev event.Event) error {
	e, ok := ev.GetEvent().(*sched.Event)
	if !ok {
		return errUnexpectedEventType
	}

	logrus.Infof("subscribed successful with ID %s", e.GetSubscribed().FrameworkId.GetValue())

	sub := e.GetSubscribed()
	connector.Instance().SetFrameworkInfoId(*sub.FrameworkId.Value)

	return s.store.UpdateFrameworkId(context.TODO(), *sub.FrameworkId.Value, nil)
}
