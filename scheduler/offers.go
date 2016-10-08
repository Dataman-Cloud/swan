package scheduler

import (
	"github.com/Sirupsen/logrus"

	mesos "github.com/Dataman-Cloud/swan/mesosproto/mesos"
	sched "github.com/Dataman-Cloud/swan/mesosproto/sched"
)

func (s *Scheduler) RequestOffers(resources []*mesos.Resource) ([]*mesos.Offer, error) {
	logrus.Info("Requesting offers...")

	var event *sched.Event

	select {
	case event = <-s.GetEvent(sched.Event_OFFERS):
	}

	if event == nil {
		call := &sched.Call{
			FrameworkId: s.framework.GetId(),
			Type:        sched.Call_REQUEST.Enum(),
			Request: &sched.Call_Request{
				Requests: []*mesos.Request{
					{
						Resources: resources,
					},
				},
			},
		}

		if _, err := s.send(call); err != nil {
			logrus.Errorf("Request offer failed: %s", err.Error())
			return nil, err
		}
		event = <-s.GetEvent(sched.Event_OFFERS)
	}

	logrus.Infof("Received %d offer(s).", len(event.Offers.Offers))
	return event.Offers.Offers, nil
}
