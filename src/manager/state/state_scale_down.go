package state

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

type StateScaleDown struct {
	Name string
	App  *App

	CurrentSlot      *Slot
	CurrentSlotIndex int
	TargetSlotIndex  int
	lock             sync.Mutex
}

func NewStateScaleDown(app *App) *StateScaleDown {
	return &StateScaleDown{
		App:  app,
		Name: APP_STATE_SCALE_DOWN,
	}
}

func (scaleDown *StateScaleDown) OnEnter() {
	logrus.Debug("state scaleDown OnEnter")

	scaleDown.App.EmitAppEvent(scaleDown.Name)

	scaleDown.CurrentSlotIndex = len(scaleDown.App.GetSlots()) - 1

	if scaleDown.App.IsFixed() {
		scaleDown.App.CurrentVersion.IP = scaleDown.App.CurrentVersion.IP[:scaleDown.CurrentSlotIndex]
	}
	scaleDown.TargetSlotIndex = int(scaleDown.App.CurrentVersion.Instances)

	scaleDown.CurrentSlot, _ = scaleDown.App.GetSlot(scaleDown.CurrentSlotIndex)
	if scaleDown.CurrentSlot != nil {
		scaleDown.CurrentSlot.KillTask()
	}
}

func (scaleDown *StateScaleDown) OnExit() {
	logrus.Debug("state scaleDown OnExit")
}

func (scaleDown *StateScaleDown) Step() {
	logrus.Debug("state scaleDown step")

	if scaleDown.SlotSafeToRemoveFromApp(scaleDown.CurrentSlot) && scaleDown.CurrentSlotIndex == scaleDown.TargetSlotIndex {
		scaleDown.App.RemoveSlot(scaleDown.CurrentSlotIndex)
		scaleDown.App.TransitTo(APP_STATE_NORMAL)
	} else if scaleDown.SlotSafeToRemoveFromApp(scaleDown.CurrentSlot) && (scaleDown.CurrentSlotIndex > scaleDown.TargetSlotIndex) {
		scaleDown.lock.Lock()

		scaleDown.App.RemoveSlot(scaleDown.CurrentSlotIndex)
		scaleDown.CurrentSlotIndex -= 1
		if scaleDown.App.IsFixed() {
			scaleDown.App.CurrentVersion.IP = scaleDown.App.CurrentVersion.IP[:scaleDown.CurrentSlotIndex]
		}
		scaleDown.CurrentSlot, _ = scaleDown.App.GetSlot(scaleDown.CurrentSlotIndex)
		scaleDown.CurrentSlot.KillTask()

		scaleDown.lock.Unlock()
	} else {
		logrus.Info("state scaleDown step, do nothing")
	}
}

func (scaleDown *StateScaleDown) SlotSafeToRemoveFromApp(slot *Slot) bool {
	return slot.StateIs(SLOT_STATE_REAP) || slot.Abnormal()
}

func (scaleDown *StateScaleDown) StateName() string {
	return scaleDown.Name
}

// state machine can transit to any state if current state is scaleDown
func (scaleDown *StateScaleDown) CanTransitTo(targetState string) bool {
	logrus.Debugf("state scaleDown CanTransitTo %s", targetState)

	return true
}
