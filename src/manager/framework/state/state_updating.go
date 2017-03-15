package state

import (
	"sync"

	"github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/src/utils"
)

var ValidNextTransitionState = []string{
	APP_STATE_CANCEL_UPDATE,
	APP_STATE_DELETING,
	APP_STATE_UPDATING,
	APP_STATE_NORMAL,
}

type StateUpdating struct {
	name string
	app  *App

	currentSlot         *Slot
	currentSlotIndex    int
	targetSlotIndex     int
	slotCountNeedUpdate int
	lock                sync.Mutex
}

func NewStateUpdating(app *App, slotCountNeedUpdate int) *StateUpdating {
	return &StateUpdating{
		app:                 app,
		name:                APP_STATE_UPDATING,
		slotCountNeedUpdate: slotCountNeedUpdate,
	}
}

func (updating *StateUpdating) OnEnter() {
	logrus.Debug("state updating OnEnter")

	updating.app.EmitAppEvent(updating.name)

	updating.currentSlotIndex = -1
	for index, slot := range updating.app.GetSlots() {
		if slot.Version == updating.app.ProposedVersion {
			updating.currentSlotIndex = index + 1
		}
	}

	if updating.currentSlotIndex == -1 {
		updating.currentSlotIndex = 0
	}
	updating.targetSlotIndex = updating.currentSlotIndex + updating.slotCountNeedUpdate - 1

	updating.currentSlot, _ = updating.app.GetSlot(updating.currentSlotIndex)
	updating.currentSlot.KillTask()
}

func (updating *StateUpdating) OnExit() {
	logrus.Debug("state updating OnExit")
}

func (updating *StateUpdating) Step() {
	logrus.Debug("state updating step")

	if (updating.currentSlot.StateIs(SLOT_STATE_REAP) ||
		updating.currentSlot.StateIs(SLOT_STATE_TASK_KILLED) ||
		updating.currentSlot.StateIs(SLOT_STATE_TASK_FINISHED) ||
		updating.currentSlot.Abnormal()) &&
		updating.currentSlotIndex <= updating.targetSlotIndex {

		logrus.Infof("archive current task")
		updating.currentSlot.Archive()
		updating.currentSlot.DispatchNewTask(updating.app.ProposedVersion)

	} else if updating.currentSlot.StateIs(SLOT_STATE_TASK_RUNNING) &&
		updating.currentSlot.Healthy() &&
		updating.currentSlotIndex < updating.targetSlotIndex {

		updating.currentSlotIndex += 1
		updating.currentSlot, _ = updating.app.GetSlot(updating.currentSlotIndex)
		updating.currentSlot.KillTask()

	} else if updating.currentSlot.StateIs(SLOT_STATE_TASK_RUNNING) &&
		updating.currentSlot.Healthy() &&
		updating.currentSlotIndex == updating.targetSlotIndex {

		if updating.currentSlotIndex == len(updating.app.GetSlots())-1 {
			logrus.Debug("state updating step, updating done,  all slots updated")

			updating.app.CurrentVersion = updating.app.ProposedVersion
			updating.app.Versions = append(updating.app.Versions, updating.app.CurrentVersion)
			updating.app.ProposedVersion = nil
			updating.app.TransitTo(APP_STATE_NORMAL)

		} else {
			logrus.Debug("state updating step, updating done,  not all slots updated")
		}
	} else {
		logrus.Info("state updating step, do nothing")
	}
}

func (updating *StateUpdating) Name() string {
	return updating.name
}

// state machine can transit to any state if current state is updating
func (updating *StateUpdating) CanTransitTo(targetState string) bool {
	logrus.Debugf("state updating CanTransitTo %s", targetState)

	if utils.SliceContains(ValidNextTransitionState, targetState) {
		return true
	}

	return false
}
