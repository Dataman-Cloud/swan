package state

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

type StateCancelUpdate struct {
	name string
	app  *App

	currentSlot      *Slot
	currentSlotIndex int
	targetSlotIndex  int
	lock             sync.Mutex
}

func NewStateCancelUpdate(app *App) *StateCancelUpdate {
	return &StateCancelUpdate{
		app:  app,
		name: APP_STATE_CANCEL_UPDATE,
	}
}

func (cancelUpdate *StateCancelUpdate) OnEnter() {
	logrus.Debug("state cancelUpdate OnEnter")

	cancelUpdate.app.EmitAppEvent(cancelUpdate.name)

	cancelUpdate.targetSlotIndex = 0
	for index, slot := range cancelUpdate.app.GetSlots() {
		if slot.Version == cancelUpdate.app.CurrentVersion {
			cancelUpdate.currentSlotIndex = index - 1
			break
		}
	}

	cancelUpdate.currentSlot, _ = cancelUpdate.app.GetSlot(cancelUpdate.currentSlotIndex)
	cancelUpdate.currentSlot.KillTask()
}

func (cancelUpdate *StateCancelUpdate) OnExit() {
	logrus.Debug("state cancelUpdate OnExit")
}

func (cancelUpdate *StateCancelUpdate) Step() {
	logrus.Debug("state cancelUpdate step")

	// when slot down but not the last one
	if (cancelUpdate.currentSlot.StateIs(SLOT_STATE_REAP) ||
		cancelUpdate.currentSlot.StateIs(SLOT_STATE_TASK_KILLED) ||
		cancelUpdate.currentSlot.Abnormal()) &&
		cancelUpdate.currentSlotIndex > cancelUpdate.targetSlotIndex {

		logrus.Infof("archive current task")
		cancelUpdate.currentSlot.Archive()
		cancelUpdate.currentSlot.DispatchNewTask(cancelUpdate.app.CurrentVersion)

		// when slot get running and pass health check
	} else if cancelUpdate.currentSlot.StateIs(SLOT_STATE_TASK_RUNNING) &&
		cancelUpdate.currentSlot.Healthy() &&
		cancelUpdate.currentSlotIndex > cancelUpdate.targetSlotIndex {

		cancelUpdate.currentSlotIndex -= 1
		cancelUpdate.currentSlot, _ = cancelUpdate.app.GetSlot(cancelUpdate.currentSlotIndex)
		cancelUpdate.currentSlot.KillTask()

		// when last slot got killed
	} else if (cancelUpdate.currentSlot.StateIs(SLOT_STATE_REAP) ||
		cancelUpdate.currentSlot.StateIs(SLOT_STATE_TASK_KILLED) ||
		cancelUpdate.currentSlot.Abnormal()) &&
		cancelUpdate.currentSlotIndex == cancelUpdate.targetSlotIndex {

		logrus.Infof("archive current task")
		cancelUpdate.currentSlot.Archive()
		cancelUpdate.currentSlot.DispatchNewTask(cancelUpdate.app.CurrentVersion)

		// when last slot got restarted
	} else if cancelUpdate.currentSlot.StateIs(SLOT_STATE_TASK_RUNNING) &&
		cancelUpdate.currentSlot.Healthy() &&
		cancelUpdate.currentSlotIndex == cancelUpdate.targetSlotIndex {
		cancelUpdate.app.ProposedVersion = nil
		cancelUpdate.app.TransitTo(APP_STATE_NORMAL)

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
