package state

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

type StateCancelUpdate struct {
	name    string
	machine *StateMachine

	currentSlot      *Slot
	currentSlotIndex int
	targetSlotIndex  int
	lock             sync.Mutex
}

func NewStateCancelUpdate(machine *StateMachine) *StateCancelUpdate {
	return &StateCancelUpdate{
		machine: machine,
		name:    APP_STATE_CANCEL_UPDATE,
	}
}

func (cancelUpdate *StateCancelUpdate) OnEnter() {
	cancelUpdate.targetSlotIndex = 0
	for index, slot := range cancelUpdate.machine.App.GetSlots() {
		if slot.Version == cancelUpdate.machine.App.CurrentVersion {
			cancelUpdate.currentSlotIndex = index - 1
			break
		}
	}

	cancelUpdate.currentSlot, _ = cancelUpdate.machine.App.GetSlot(cancelUpdate.currentSlotIndex)

	cancelUpdate.currentSlot.Archive()
	cancelUpdate.currentSlot.UpdateTask(cancelUpdate.machine.App.CurrentVersion)
}

func (cancelUpdate *StateCancelUpdate) OnExit() {
	logrus.Debug("state cancelUpdate OnExit")
}

func (cancelUpdate *StateCancelUpdate) Step() {
	logrus.Debug("state cancelUpdate step")

	if (cancelUpdate.currentSlot.StateIs(SLOT_STATE_REAP) ||
		cancelUpdate.currentSlot.StateIs(SLOT_STATE_TASK_KILLED) ||
		cancelUpdate.currentSlot.Abnormal()) &&
		cancelUpdate.currentSlotIndex > cancelUpdate.targetSlotIndex {

		logrus.Infof("archive current task")
		cancelUpdate.currentSlot.Archive()

		cancelUpdate.currentSlotIndex -= 1
		cancelUpdate.currentSlot, _ = cancelUpdate.machine.App.GetSlot(cancelUpdate.currentSlotIndex)
		cancelUpdate.currentSlot.UpdateTask(cancelUpdate.machine.App.CurrentVersion)

	} else if (cancelUpdate.currentSlot.StateIs(SLOT_STATE_REAP) ||
		cancelUpdate.currentSlot.StateIs(SLOT_STATE_TASK_KILLED) ||
		cancelUpdate.currentSlot.Abnormal()) &&
		cancelUpdate.currentSlotIndex == cancelUpdate.targetSlotIndex {

		logrus.Infof("archive current task")
		cancelUpdate.machine.App.StateMachine.TransitTo(APP_STATE_NORMAL)
	} else {
		logrus.Info("state cancelUpdate step, do nothing")
	}
}

func (cancelUpdate *StateCancelUpdate) Name() string {
	return cancelUpdate.name
}

// state machine can transit to any state if current state is cancelUpdate
func (cancelUpdate *StateCancelUpdate) CanTransitTo(targetState string) bool {
	logrus.Debugf("state cancelUpdate CanTransitTo %s", targetState)

	if targetState == APP_STATE_DELETING {
		return true
	}

	if targetState == APP_STATE_NORMAL {
		return true
	}

	return false
}
