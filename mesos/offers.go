package mesos

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/Dataman-Cloud/swan/proto/mesos"
	"github.com/Dataman-Cloud/swan/proto/sched"
)

type resource struct {
	cpus     float64 `json:"cpus"`
	mem      float64 `json:"mem"`
	disk     float64 `json:"disk"`
	portsNum int     `json:"portsNum"`
}

var requestOfferTimeout = time.Second * time.Duration(300)

//func (s *Scheduler) RequestOffer(res *resource) (*mesos.Offer, error) {
//	log.Info("Requesting offers")
//
//	//for {
//	//	select {
//	//	case agent := <-s.selectAgentForTask(res):
//	//		if agent == nil {
//	//			continue
//	//		}
//
//	//		//return agent.getOffers()[0], nil
//	//		return nil, nil
//	//	case <-time.After(requestOfferTimeout):
//	//		return nil, errors.New("Offer timeout")
//	//	}
//	//}
//}

func (s *Scheduler) DeclineOffer(offer *mesos.Offer) error {
	call := &sched.Call{
		FrameworkId: s.FrameworkId(),
		Type:        sched.Call_DECLINE.Enum(),
		Decline: &sched.Call_Decline{
			OfferIds: []*mesos.OfferID{
				{
					Value: offer.GetId().Value,
				},
			},
			Filters: &mesos.Filters{
				RefuseSeconds: proto.Float64(1),
			},
		},
	}

	resp, err := s.Send(call)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("%d", resp.StatusCode)
	}

	return nil
}
