package engine

import (
	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
)

func OfferHandler(h *Handler) (*Handler, error) {
	logrus.WithFields(logrus.Fields{"handler": "offer"}).Debugf("")

	for _, offer := range h.MesosEvent.Event.Offers.Offers {
		// when no pending offer slot
		if len(h.EngineRef.Allocator.PendingOfferSlots) == 0 {

			RejectOffer(h, offer)
		} else {

			offerWrapper := state.NewOfferWrapper(offer)
			taskInfos := make([]*mesos.TaskInfo, 0)
			for {
				// loop through all pending offer slots
				slot := h.EngineRef.Allocator.NextPendingOffer()
				if slot == nil {
					break
				}

				match := slot.TestOfferMatch(offerWrapper)
				if match {
					slot.SetState(state.SLOT_STATE_TASK_DISPATCHED)

					// TODO the following code logic complex, need improvement
					// offerWrapper cpu/mem/disk deduction recorded within the obj itself
					_, taskInfo := slot.ReserveOfferAndPrepareTaskInfo(offerWrapper)
					h.EngineRef.Allocator.SetOfferIdForSlotName(offer.GetId(), slot.Name)
					taskInfos = append(taskInfos, taskInfo)

				} else {
					// put the slot back into the queue, in the end
					h.EngineRef.Allocator.PutSlotBackToPendingQueue(slot)
				}
			}
			LaunchTaskInfos(h, offer, taskInfos)
		}
	}

	return h, nil
}

func LaunchTaskInfos(h *Handler, offer *mesos.Offer, taskInfos []*mesos.TaskInfo) {
	call := &sched.Call{
		FrameworkId: h.EngineRef.Scheduler.Framework.GetId(),
		Type:        sched.Call_ACCEPT.Enum(),
		Accept: &sched.Call_Accept{
			OfferIds: []*mesos.OfferID{
				offer.GetId(),
			},
			Operations: []*mesos.Offer_Operation{
				&mesos.Offer_Operation{
					Type: mesos.Offer_Operation_LAUNCH.Enum(),
					Launch: &mesos.Offer_Operation_Launch{
						TaskInfos: taskInfos,
					},
				},
			},
			Filters: &mesos.Filters{RefuseSeconds: proto.Float64(1)},
		},
	}

	h.Response.Calls = append(h.Response.Calls, call)
}

func RejectOffer(h *Handler, offer *mesos.Offer) {
	call := &sched.Call{
		FrameworkId: h.EngineRef.Scheduler.Framework.GetId(),
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

	h.Response.Calls = append(h.Response.Calls, call)
}
