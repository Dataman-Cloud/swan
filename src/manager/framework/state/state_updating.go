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
	name    string
	machine *StateMachine

	currentSlot         *Slot
	currentSlotIndex    int
	targetSlotIndex     int
	slotCountNeedUpdate int
	lock                sync.Mutex
}

func NewStateUpdating(machine *StateMachine, slotCountNeedUpdate int) *StateUpdating {
	return &StateUpdating{
		machine:             machine,
		name:                APP_STATE_UPDATING,
		slotCountNeedUpdate: slotCountNeedUpdate,
	}
}

func (updating *StateUpdating) OnEnter() {
	updating.currentSlotIndex = -1
	for index, slot := range updating.machine.App.GetSlots() {
		if slot.Version == updating.machine.App.ProposedVersion {
			updating.currentSlotIndex = index + 1
		}
	}

	if updating.currentSlotIndex == -1 {
		updating.currentSlotIndex = 0
	}
	updating.targetSlotIndex = updating.currentSlotIndex + updating.slotCountNeedUpdate - 1

	updating.currentSlot, _ = updating.machine.App.GetSlot(updating.currentSlotIndex)
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
		updating.currentSlot.DispatchNewTask(updating.machine.App.ProposedVersion)

	} else if updating.currentSlot.StateIs(SLOT_STATE_TASK_RUNNING) &&
		updating.currentSlot.Healthy() &&
		updating.currentSlotIndex < updating.targetSlotIndex {

		updating.currentSlotIndex += 1
		updating.currentSlot, _ = updating.machine.App.GetSlot(updating.currentSlotIndex)
		updating.currentSlot.KillTask()

	} else if updating.currentSlot.StateIs(SLOT_STATE_TASK_RUNNING) &&
		updating.currentSlot.Healthy() &&
		updating.currentSlotIndex == updating.targetSlotIndex {

		if updating.currentSlotIndex == len(updating.machine.App.GetSlots())-1 {
			logrus.Debug("state updating step, updating done,  all slots updated")

			updating.machine.App.CurrentVersion = updating.machine.App.ProposedVersion
			updating.machine.App.Versions = append(updating.machine.App.Versions, updating.machine.App.CurrentVersion)
			updating.machine.App.ProposedVersion = nil
			updating.machine.TransitTo(APP_STATE_NORMAL)

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
