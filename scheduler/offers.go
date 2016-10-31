package scheduler

import (
	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
)

func (s *Scheduler) RequestOffers(resources []*mesos.Resource) ([]*mesos.Offer, error) {
	logrus.Info("Requesting offers")

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

// DeclineResource is used to send DECLINE request to mesos to release offer. This
// is very important, otherwise resource will be taked until framework exited.
func (s *Scheduler) DeclineResource(offerId *string) {
	call := &sched.Call{
		FrameworkId: s.framework.GetId(),
		Type:        sched.Call_DECLINE.Enum(),
		Decline: &sched.Call_Decline{
			OfferIds: []*mesos.OfferID{
				{
					Value: offerId,
				},
			},
			Filters: &mesos.Filters{
				RefuseSeconds: proto.Float64(1),
			},
		},
	}

	s.send(call)
}

func (s *Scheduler) OfferedResources(offer *mesos.Offer) (cpus, mem, disk float64) {
	for _, res := range offer.GetResources() {
		if res.GetName() == "cpus" {
			cpus += *res.GetScalar().Value
		}
		if res.GetName() == "mem" {
			mem += *res.GetScalar().Value
		}
		if res.GetName() == "disk" {
			disk += *res.GetScalar().Value
		}
	}

	return
}
