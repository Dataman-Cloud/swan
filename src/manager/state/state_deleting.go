package state

import (
	"sync"

	"github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/src/manager/event"
)

type StateDeleting struct {
	Name string
	App  *App

	CurrentSlot      *Slot
	CurrentSlotIndex int
	TargetSlotIndex  int
	lock             sync.Mutex
}

func NewStateDeleting(app *App) *StateDeleting {
	return &StateDeleting{
		App:  app,
		Name: APP_STATE_DELETING,
	}
}

func (deleting *StateDeleting) OnEnter() {
	logrus.Debug("state deleting OnEnter")

	deleting.App.EmitAppEvent(deleting.Name)

	deleting.CurrentSlotIndex = len(deleting.App.GetSlots()) - 1
	deleting.TargetSlotIndex = 0

	deleting.CurrentSlot, _ = deleting.App.GetSlot(deleting.CurrentSlotIndex)
	if deleting.CurrentSlot != nil {
		deleting.CurrentSlot.KillTask()
	} else {
		deleting.App.Remove() // remove self from boltdb store

		deleting.App.UserEventChan <- &event.UserEvent{ // signal scheduler in-memory store to remove this app
			Type:  event.EVENT_TYPE_USER_INVALID_APPS,
			Param: deleting.App.ID,
		}
	}
}

func (deleting *StateDeleting) OnExit() {
	logrus.Debug("state deleting OnExit")
}

func (deleting *StateDeleting) Step() {
	logrus.Debug("state deleting step")

	if deleting.SlotSafeToRemoveFromApp(deleting.CurrentSlot) && deleting.CurrentSlotIndex == deleting.TargetSlotIndex {
		deleting.App.RemoveSlot(deleting.CurrentSlotIndex)
		deleting.App.Remove() // remove self from boltdb store

		deleting.App.UserEventChan <- &event.UserEvent{ // signal scheduler in-memory store to remove this app
			Type:  event.EVENT_TYPE_USER_INVALID_APPS,
			Param: deleting.App.ID,
		}
	} else if deleting.SlotSafeToRemoveFromApp(deleting.CurrentSlot) && (deleting.CurrentSlotIndex > deleting.TargetSlotIndex) {
		deleting.lock.Lock()

		deleting.App.RemoveSlot(deleting.CurrentSlotIndex)
		deleting.CurrentSlotIndex -= 1
		deleting.CurrentSlot, _ = deleting.App.GetSlot(deleting.CurrentSlotIndex)
		if deleting.CurrentSlot != nil {
			deleting.CurrentSlot.KillTask()
		}

		deleting.lock.Unlock()
	} else {
		logrus.Info("state deleting step, do nothing")
	}

}

func (deleting *StateDeleting) SlotSafeToRemoveFromApp(slot *Slot) bool {
	return slot.StateIs(SLOT_STATE_REAP) || slot.Abnormal()
}

func (deleting *StateDeleting) StateName() string {
	return deleting.Name
}

// state machine can transit to any state if current state is deleting
func (deleting *StateDeleting) CanTransitTo(targetState string) bool {
	logrus.Debugf("state deleting CanTransitTo %s", targetState)

	if targetState == APP_STATE_DELETING {
		return true
	}

	return false
}
