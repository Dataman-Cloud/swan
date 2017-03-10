package state

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

type StateScaleUp struct {
	name    string
	machine *StateMachine

	currentSlot      *Slot
	currentSlotIndex int
	targetSlotIndex  int
	lock             sync.Mutex
}

func NewStateScaleUp(machine *StateMachine) *StateScaleUp {
	return &StateScaleUp{
		machine: machine,
		name:    APP_STATE_SCALE_UP,
	}
}

func (scaleUp *StateScaleUp) OnEnter() {
	scaleUp.currentSlotIndex = len(scaleUp.machine.App.GetSlots()) - 1
	scaleUp.targetSlotIndex = int(scaleUp.machine.App.CurrentVersion.Instances) - 1

	scaleUp.currentSlot = NewSlot(scaleUp.machine.App, scaleUp.machine.App.CurrentVersion, scaleUp.currentSlotIndex)
	scaleUp.machine.App.SetSlot(scaleUp.currentSlotIndex, scaleUp.currentSlot)
	scaleUp.currentSlot.DispatchNewTask(scaleUp.currentSlot.Version)
}

func (scaleUp *StateScaleUp) OnExit() {
	logrus.Debug("state scaleUp OnExit")
}

func (scaleUp *StateScaleUp) Step() {
	logrus.Debug("state scaleUp step")

	if scaleUp.currentSlotIndex == scaleUp.targetSlotIndex && scaleUp.currentSlot.Healthy() {
		scaleUp.machine.TransitTo(APP_STATE_NORMAL)
	} else if scaleUp.currentSlot.Healthy() && scaleUp.currentSlotIndex < scaleUp.targetSlotIndex {
		scaleUp.lock.Lock()

		scaleUp.currentSlotIndex += 1
		scaleUp.currentSlot = NewSlot(scaleUp.machine.App, scaleUp.machine.App.CurrentVersion, scaleUp.currentSlotIndex)
		scaleUp.machine.App.SetSlot(scaleUp.currentSlotIndex, scaleUp.currentSlot)
		scaleUp.currentSlot.DispatchNewTask(scaleUp.currentSlot.Version)

		scaleUp.lock.Unlock()
	} else {
		logrus.Info("state scaleUp step, do nothing")
	}
}

func (scaleUp *StateScaleUp) Name() string {
	return scaleUp.name
}

// state machine can transit to any state if current state is scaleUp
func (scaleUp *StateScaleUp) CanTransitTo(targetState string) bool {
	logrus.Debugf("state scaleUp CanTransitTo %s", targetState)

	return true
}
