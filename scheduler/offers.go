package scheduler

import (
	"fmt"
	"github.com/Sirupsen/logrus"

	mesos "github.com/Dataman-Cloud/swan/mesosproto/mesos"
	sched "github.com/Dataman-Cloud/swan/mesosproto/sched"
)

func (s *Scheduler) RequestOffer(resources []*mesos.Resource) (*mesos.Offer, error) {
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

	logrus.Info("Pickup satisfied offer")
	for _, offer := range event.Offers.Offers {
		cpus, mems := s.offeredResources(offer)
		cpusNeeded, memsNeeded := s.neededResource(resources)
		if cpus >= cpusNeeded && mems >= memsNeeded {
			logrus.WithFields(
				logrus.Fields{"offerID=": *offer.Id.Value, "hostname=": *offer.Hostname},
			).Info("Offer satisfied")
			return offer, nil
		}
	}

	return nil, fmt.Errorf("All offers are not satisfied!")
}

// offeredResources
func (s *Scheduler) offeredResources(offer *mesos.Offer) (cpus, mems float64) {
	for _, res := range offer.GetResources() {
		if res.GetName() == "cpus" {
			cpus += *res.GetScalar().Value
		}
		if res.GetName() == "mem" {
			mems += *res.GetScalar().Value
		}
	}

	return
}

func (s *Scheduler) neededResource(resources []*mesos.Resource) (cpusNeeded, memsNeeded float64) {
	for _, res := range resources {
		if res.GetName() == "cpus" {
			cpusNeeded = *res.GetScalar().Value
		}

		if res.GetName() == "mem" {
			memsNeeded = *res.GetScalar().Value
		}
	}

	return
}
