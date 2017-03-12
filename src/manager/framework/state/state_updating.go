package state

import (
	"sync"

	"github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/src/utils"
)

var ValidNextTransitionState = []string{
	APP_STATE_CANCEL_UPDATE,
	APP_STATE_DELETING,
}

type StateUpdating struct {
	name    string
	machine *StateMachine

	currentSlot      *Slot
	currentSlotIndex int
	lock             sync.Mutex
}

func NewStateUpdating(machine *StateMachine) *StateUpdating {
	return &StateUpdating{
		machine: machine,
		name:    APP_STATE_UPDATING,
	}
}

func (updating *StateUpdating) OnEnter() {
	updating.currentSlotIndex = 0

	updating.currentSlot, _ = updating.machine.App.GetSlot(updating.currentSlotIndex)
	updating.machine.App.SetSlot(updating.currentSlotIndex, updating.currentSlot)
	updating.currentSlot.UpdateTask(updating.machine.App.ProposedVersion)
}

func (updating *StateUpdating) OnExit() {
	logrus.Debug("state updating OnExit")
}

func (updating *StateUpdating) Step() {
	logrus.Debug("state updating step")

	if updating.currentSlot.StateIs(SLOT_STATE_REAP) ||
		updating.currentSlot.StateIs(SLOT_STATE_TASK_KILLED) ||
		updating.currentSlot.Abnormal() {

		logrus.Infof("archive current task")
		updating.currentSlot.Archive()
		updating.currentSlot.DispatchNewTask(updating.currentSlot.Version)
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

	if targetState == APP_STATE_DELETING {
		return true
	}

	return false
}
