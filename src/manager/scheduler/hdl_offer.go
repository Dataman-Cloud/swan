package scheduler

import (
	"github.com/Dataman-Cloud/swan/src/manager/connector"
	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/state"
	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"

	"github.com/golang/protobuf/proto"
)

func OfferHandler(s *Scheduler, ev event.Event) error {
	e, ok := ev.GetEvent().(*sched.Event)
	if !ok {
		return errUnexpectedEventType
	}

	for _, offer := range e.Offers.Offers {
		// when no pending offer slot
		offerWrapper := state.NewOfferWrapper(offer)
		taskInfos := make([]*mesos.TaskInfo, 0)
		nonMatchedSlots := make([]*state.Slot, 0)
		for {
			// loop through all pending offer slots
			slot := state.OfferAllocatorInstance().ShiftNextPendingOffer()
			if slot == nil {
				break
			}

			match := slot.TestOfferMatch(offerWrapper)
			if match {
				// TODO the following code logic complex, need improvement
				// offerWrapper cpu/mem/disk deduction recorded within the obj itself
				_, taskInfo := slot.ReserveOfferAndPrepareTaskInfo(offerWrapper)
				state.OfferAllocatorInstance().SetOfferSlotMap(offer, slot)
				taskInfos = append(taskInfos, taskInfo)
			} else {
				// put the slot back into the queue, in the end
				nonMatchedSlots = append(nonMatchedSlots, slot)
			}
		}

		for _, slot := range nonMatchedSlots {
			state.OfferAllocatorInstance().PutSlotBackToPendingQueue(slot)
		}

		if len(taskInfos) > 0 {
			LaunchTaskInfos(offer, taskInfos)
		} else { // reject offer here
			RejectOffer(offer)
		}
	}

	return nil
}

func LaunchTaskInfos(offer *mesos.Offer, taskInfos []*mesos.TaskInfo) {
	call := &sched.Call{
		FrameworkId: connector.Instance().FrameworkInfo.GetId(),
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

	connector.Instance().SendCall(call)
}

func RejectOffer(offer *mesos.Offer) {
	call := &sched.Call{
		FrameworkId: connector.Instance().FrameworkInfo.GetId(),
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

	connector.Instance().SendCall(call)
}
