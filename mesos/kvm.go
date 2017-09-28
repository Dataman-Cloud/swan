package mesos

import (
	"fmt"
	"net/http"
	"strings"

	magent "github.com/Dataman-Cloud/swan/mesos/agent"
	"github.com/Dataman-Cloud/swan/mesosproto"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
)

// launch grouped runtime tasks with specified mesos offers
func (s *Scheduler) launchGroupKvmTasksWithOffers(offers []*magent.Offer, tasks []*Task) error {
	var (
		appId = strings.SplitN(tasks[0].GetName(), ".", 2)[1]
	)

	// build each tasks: set mesos agent id & build
	for _, task := range tasks {
		task.AgentId = &mesosproto.AgentID{
			Value: proto.String(offers[0].GetAgentId()),
		}
		task.Build()
	}

	// memo update each db tasks' AgentID, IP ...
	for _, t := range tasks {
		dbtask, err := s.db.GetKvmTask(appId, t.GetTaskId().GetValue())
		if err != nil {
			continue
		}

		dbtask.AgentId = t.AgentId.GetValue()
		dbtask.IP = offers[0].GetHostname()

		if err := s.db.UpdateKvmTask(appId, dbtask); err != nil {
			log.Errorln("update db kvm task got error: %v", err)
			continue
		}
	}

	// Construct Mesos Launch Call
	var (
		offerIds  = []*mesosproto.OfferID{}
		taskInfos = []*mesosproto.TaskInfo{}
	)

	for _, offer := range offers {
		offerIds = append(offerIds, &mesosproto.OfferID{
			Value: proto.String(offer.GetId()),
		})
	}

	for _, task := range tasks {
		taskInfos = append(taskInfos, &task.TaskInfo)
	}

	call := &mesosproto.Call{
		FrameworkId: s.FrameworkId(),
		Type:        mesosproto.Call_ACCEPT.Enum(),
		Accept: &mesosproto.Call_Accept{
			OfferIds: offerIds,
			Operations: []*mesosproto.Offer_Operation{
				&mesosproto.Offer_Operation{
					Type: mesosproto.Offer_Operation_LAUNCH.Enum(),
					Launch: &mesosproto.Offer_Operation_Launch{
						TaskInfos: taskInfos,
					},
				},
			},
			Filters: &mesosproto.Filters{RefuseSeconds: proto.Float64(1)},
		},
	}

	log.Printf("Launching %d kvm task(s) on agent %s", len(tasks), offers[0].GetHostname())

	// send call
	if _, err := s.SendCall(call, http.StatusAccepted); err != nil {
		log.Errorln("launch().SendCall() error:", err)
		return fmt.Errorf("send launch call got error: %v", err)
	}

	return nil
}
