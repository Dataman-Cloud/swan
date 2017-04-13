package state

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

type StateCancelUpdate struct {
	Name string
	App  *App

	CurrentSlot      *Slot
	CurrentSlotIndex int
	TargetSlotIndex  int
	lock             sync.Mutex
}

func NewStateCancelUpdate(app *App) *StateCancelUpdate {
	return &StateCancelUpdate{
		App:  app,
		Name: APP_STATE_CANCEL_UPDATE,
	}
}

func (cancelUpdate *StateCancelUpdate) OnEnter() {
	logrus.Debug("state cancelUpdate OnEnter")

	cancelUpdate.App.EmitAppEvent(cancelUpdate.Name)

	cancelUpdate.TargetSlotIndex = 0
	for index, slot := range cancelUpdate.App.GetSlots() {
		if slot.Version.ID == cancelUpdate.App.CurrentVersion.ID {
			cancelUpdate.CurrentSlotIndex = index - 1
			break
		}
	}

	cancelUpdate.CurrentSlot, _ = cancelUpdate.App.GetSlot(cancelUpdate.CurrentSlotIndex)
	cancelUpdate.CurrentSlot.KillTask()
}

func (cancelUpdate *StateCancelUpdate) OnExit() {
	logrus.Debug("state cancelUpdate OnExit")
}

func (cancelUpdate *StateCancelUpdate) Step() {
	logrus.Debug("state cancelUpdate step")

	// when slot down but not the last one
	if (cancelUpdate.CurrentSlot.StateIs(SLOT_STATE_REAP) ||
		cancelUpdate.CurrentSlot.Abnormal()) &&
		cancelUpdate.CurrentSlotIndex > cancelUpdate.TargetSlotIndex {

		logrus.Infof("archive current task")
		cancelUpdate.CurrentSlot.Archive()
		cancelUpdate.CurrentSlot.DispatchNewTask(cancelUpdate.App.CurrentVersion)

		// when slot get running and pass health check
	} else if cancelUpdate.CurrentSlot.StateIs(SLOT_STATE_TASK_RUNNING) &&
		cancelUpdate.CurrentSlot.Healthy() &&
		cancelUpdate.CurrentSlotIndex > cancelUpdate.TargetSlotIndex {

		cancelUpdate.CurrentSlotIndex -= 1
		cancelUpdate.CurrentSlot, _ = cancelUpdate.App.GetSlot(cancelUpdate.CurrentSlotIndex)
		cancelUpdate.CurrentSlot.KillTask()

		// when last slot got killed
	} else if (cancelUpdate.CurrentSlot.StateIs(SLOT_STATE_REAP) ||
		cancelUpdate.CurrentSlot.Abnormal()) &&
		cancelUpdate.CurrentSlotIndex == cancelUpdate.TargetSlotIndex {

		logrus.Infof("archive current task")
		cancelUpdate.CurrentSlot.Archive()
		cancelUpdate.CurrentSlot.DispatchNewTask(cancelUpdate.App.CurrentVersion)

		// when last slot got restarted
	} else if cancelUpdate.CurrentSlot.StateIs(SLOT_STATE_TASK_RUNNING) &&
		cancelUpdate.CurrentSlot.Healthy() &&
		cancelUpdate.CurrentSlotIndex == cancelUpdate.TargetSlotIndex {
		cancelUpdate.App.ProposedVersion = nil
		cancelUpdate.App.TransitTo(APP_STATE_NORMAL)

	} else {
		logrus.Info("state cancelUpdate step, do nothing")
	}
}

func (cancelUpdate *StateCancelUpdate) StateName() string {
	return cancelUpdate.Name
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
