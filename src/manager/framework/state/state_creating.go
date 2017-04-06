package state

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

type StateCreating struct {
	Name string
	App  *App

	CurrentSlot      *Slot
	CurrentSlotIndex int
	TargetSlotIndex  int
	lock             sync.Mutex
}

func NewStateCreating(app *App) *StateCreating {
	return &StateCreating{
		App:  app,
		Name: APP_STATE_CREATING,
	}
}

func (creating *StateCreating) OnEnter() {
	logrus.Debug("state creating OnEnter")

	creating.App.EmitAppEvent(creating.Name)

	creating.CurrentSlotIndex = 0
	creating.TargetSlotIndex = int(creating.App.CurrentVersion.Instances) - 1

	creating.CurrentSlot = NewSlot(creating.App, creating.App.CurrentVersion, creating.CurrentSlotIndex)
	creating.App.SetSlot(creating.CurrentSlotIndex, creating.CurrentSlot)
	creating.CurrentSlot.DispatchNewTask(creating.CurrentSlot.Version)
}

func (creating *StateCreating) OnExit() {
	logrus.Debug("state creating OnExit")
}

func (creating *StateCreating) Step() {
	logrus.Debug("state creating step")

	if creating.CurrentSlotIndex == creating.TargetSlotIndex && creating.CurrentSlot.Healthy() {
		creating.App.TransitTo(APP_STATE_NORMAL)
	} else if creating.CurrentSlot.Healthy() && creating.CurrentSlotIndex < creating.TargetSlotIndex {
		creating.lock.Lock()

		creating.CurrentSlotIndex += 1
		creating.CurrentSlot = NewSlot(creating.App, creating.App.CurrentVersion, creating.CurrentSlotIndex)
		creating.App.SetSlot(creating.CurrentSlotIndex, creating.CurrentSlot)
		creating.CurrentSlot.DispatchNewTask(creating.CurrentSlot.Version)

		creating.lock.Unlock()
	} else {
		logrus.Debug("state creating step, do nothing")
	}
}

func (creating *StateCreating) StateName() string {
	return creating.Name
}

// state machine can transit to any state if current state is creating
func (creating *StateCreating) CanTransitTo(targetState string) bool {
	logrus.Debugf("state creating CanTransitTo %s", targetState)

	return true
}
