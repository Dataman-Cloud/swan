package scheduler

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gogo/protobuf/proto"

	mesos "github.com/Dataman-Cloud/swan/mesosproto/mesos"
	sched "github.com/Dataman-Cloud/swan/mesosproto/sched"
)

func (s *Scheduler) RequestOffer(resources []*mesos.Resource) (*mesos.Offer, error) {
	logrus.Info("Requesting offers")

	var event *sched.Event
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

	logrus.Infof("Received %d offer(s).", len(event.Offers.Offers))
	for _, offer := range event.Offers.Offers {
		cpus, mem, disk := s.offeredResources(offer)
		cpusNeeded, memNeeded, diskNeeded := s.neededResource(resources)
		if cpus >= cpusNeeded && mem >= memNeeded && disk >= diskNeeded {
			logrus.WithFields(
				logrus.Fields{"cpus": cpus, "mem": mem, "disk": disk},
			).Info("Satisfied offer")
			return offer, nil
		}
		s.DeclineResource(offer.GetId().Value)
	}

	return nil, fmt.Errorf("All offers are not satisfied!")
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

// offeredResources
func (s *Scheduler) offeredResources(offer *mesos.Offer) (cpus, mem, disk float64) {
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

func (s *Scheduler) neededResource(resources []*mesos.Resource) (cpusNeeded, memNeeded, diskNeeded float64) {
	for _, res := range resources {
		if res.GetName() == "cpus" {
			cpusNeeded = *res.GetScalar().Value
		}

		if res.GetName() == "mem" {
			memNeeded = *res.GetScalar().Value
		}

		if res.GetName() == "disk" {
			diskNeeded = *res.GetScalar().Value
		}
	}

	return
}
