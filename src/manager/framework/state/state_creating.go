package state

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

type StateCreating struct {
	name    string
	machine *StateMachine

	currentSlot      *Slot
	currentSlotIndex int
	targetSlotIndex  int
	lock             sync.Mutex
}

func NewStateCreating(machine *StateMachine) *StateCreating {
	return &StateCreating{
		machine: machine,
		name:    APP_STATE_CREATING,
	}
}

func (creating *StateCreating) OnEnter() {
	creating.currentSlotIndex = 0
	creating.targetSlotIndex = int(creating.machine.App.CurrentVersion.Instances)

	creating.currentSlot = NewSlot(creating.machine.App, creating.machine.App.CurrentVersion, creating.currentSlotIndex)
	creating.machine.App.SetSlot(creating.currentSlotIndex, creating.currentSlot)
	creating.currentSlot.DispatchNewTask(creating.currentSlot.Version)
}

func (creating *StateCreating) OnExit() {
	logrus.Debug("state creating OnExit")
}

func (creating *StateCreating) Step() {
	logrus.Debug("state creating step")

	if creating.currentSlotIndex == creating.targetSlotIndex && creating.currentSlot.Healthy() {
		creating.machine.TransitTo(APP_STATE_NORMAL)
	} else if creating.currentSlot.Healthy() && creating.currentSlotIndex < creating.targetSlotIndex {
		creating.lock.Lock()

		creating.currentSlotIndex += 1
		creating.currentSlot = NewSlot(creating.machine.App, creating.machine.App.CurrentVersion, creating.currentSlotIndex)
		creating.machine.App.SetSlot(creating.currentSlotIndex, creating.currentSlot)
		creating.currentSlot.DispatchNewTask(creating.currentSlot.Version)

		creating.lock.Unlock()
	} else {
		logrus.Info("state creating step, do nothing")
	}
}

func (creating *StateCreating) Name() string {
	return creating.name
}

// state machine can transit to any state if current state is creating
func (creating *StateCreating) CanTransitTo(targetState string) bool {
	logrus.Debugf("state creating CanTransitTo %s", targetState)

	return true
}
