package state

import (
	"sync"

	"github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
)

type StateDeleting struct {
	name    string
	machine *StateMachine

	currentSlot      *Slot
	currentSlotIndex int
	targetSlotIndex  int
	lock             sync.Mutex
}

func NewStateDeleting(machine *StateMachine) *StateDeleting {
	return &StateDeleting{
		machine: machine,
		name:    APP_STATE_DELETING,
	}
}

func (deleting *StateDeleting) OnEnter() {
	logrus.Debug("state deleting OnEnter")

	deleting.currentSlotIndex = len(deleting.machine.App.GetSlots()) - 1
	deleting.targetSlotIndex = 0

	deleting.currentSlot, _ = deleting.machine.App.GetSlot(deleting.currentSlotIndex)
	deleting.currentSlot.KillTask()
}

func (deleting *StateDeleting) OnExit() {
	logrus.Debug("state deleting OnExit")
}

func (deleting *StateDeleting) Step() {
	logrus.Debug("state deleting step")

	if deleting.SlotSafeToRemoveFromApp(deleting.currentSlot) && deleting.currentSlotIndex == deleting.targetSlotIndex {
		deleting.machine.App.RemoveSlot(deleting.currentSlotIndex)

		deleting.machine.App.Remove() // remove self from boltdb store

		deleting.machine.App.UserEventChan <- &event.UserEvent{ // signal scheduler in-memory store to remove this app
			Type:  event.EVENT_TYPE_USER_INVALID_APPS,
			Param: deleting.machine.App.ID,
		}
	} else if deleting.SlotSafeToRemoveFromApp(deleting.currentSlot) && (deleting.currentSlotIndex > deleting.targetSlotIndex) {
		deleting.lock.Lock()

		deleting.machine.App.RemoveSlot(deleting.currentSlotIndex)
		deleting.currentSlotIndex -= 1
		deleting.currentSlot, _ = deleting.machine.App.GetSlot(deleting.currentSlotIndex)
		deleting.currentSlot.KillTask()

		deleting.lock.Unlock()
	} else {
		logrus.Info("state deleting step, do nothing")
	}
}

func (deleting *StateDeleting) SlotSafeToRemoveFromApp(slot *Slot) bool {
	return slot.StateIs(SLOT_STATE_REAP) || slot.Abnormal()
}

func (deleting *StateDeleting) Name() string {
	return deleting.name
}

// state machine can transit to any state if current state is deleting
func (deleting *StateDeleting) CanTransitTo(targetState string) bool {
	logrus.Debugf("state deleting CanTransitTo %s", targetState)

	return false
}
