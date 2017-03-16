package state

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

type StateScaleUp struct {
	Name string
	App  *App

	CurrentSlot      *Slot
	CurrentSlotIndex int
	TargetSlotIndex  int
	lock             sync.Mutex
}

func NewStateScaleUp(app *App) *StateScaleUp {
	return &StateScaleUp{
		App:  app,
		Name: APP_STATE_SCALE_UP,
	}

}
func (scaleUp *StateScaleUp) OnEnter() {
	logrus.Debug("state scaleUp OnEnter")

	scaleUp.App.EmitAppEvent(scaleUp.Name)

	scaleUp.CurrentSlotIndex = len(scaleUp.App.GetSlots())
	scaleUp.TargetSlotIndex = int(scaleUp.App.CurrentVersion.Instances) - 1

	scaleUp.CurrentSlot = NewSlot(scaleUp.App, scaleUp.App.CurrentVersion, scaleUp.CurrentSlotIndex)
	scaleUp.App.SetSlot(scaleUp.CurrentSlotIndex, scaleUp.CurrentSlot)
	scaleUp.CurrentSlot.DispatchNewTask(scaleUp.CurrentSlot.Version)
}

func (scaleUp *StateScaleUp) OnExit() {
	logrus.Debug("state scaleUp OnExit")
}

func (scaleUp *StateScaleUp) Step() {
	logrus.Debug("state scaleUp step")

	if scaleUp.CurrentSlotIndex == scaleUp.TargetSlotIndex && scaleUp.CurrentSlot.Healthy() {
		scaleUp.App.TransitTo(APP_STATE_NORMAL)
	} else if scaleUp.CurrentSlot.Healthy() && scaleUp.CurrentSlotIndex < scaleUp.TargetSlotIndex {
		scaleUp.lock.Lock()

		scaleUp.CurrentSlotIndex += 1
		scaleUp.CurrentSlot = NewSlot(scaleUp.App, scaleUp.App.CurrentVersion, scaleUp.CurrentSlotIndex)
		scaleUp.App.SetSlot(scaleUp.CurrentSlotIndex, scaleUp.CurrentSlot)
		scaleUp.CurrentSlot.DispatchNewTask(scaleUp.CurrentSlot.Version)

		scaleUp.lock.Unlock()
	} else {
		logrus.Info("state scaleUp step, do nothing")
	}
}

func (scaleUp *StateScaleUp) StateName() string {
	return scaleUp.Name
}

// state machine can transit to any state if current state is scaleUp
func (scaleUp *StateScaleUp) CanTransitTo(targetState string) bool {
	logrus.Debugf("state scaleUp CanTransitTo %s", targetState)

	return true
}
