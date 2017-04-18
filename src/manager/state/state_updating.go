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
	Name string
	App  *App

	CurrentSlot         *Slot
	CurrentSlotIndex    int
	TargetSlotIndex     int
	SlotCountNeedUpdate int
	lock                sync.Mutex
}

func NewStateUpdating(app *App, slotCountNeedUpdate int) *StateUpdating {
	return &StateUpdating{
		App:                 app,
		Name:                APP_STATE_UPDATING,
		SlotCountNeedUpdate: slotCountNeedUpdate,
	}
}

func (updating *StateUpdating) OnEnter() {
	logrus.Debug("state updating OnEnter")

	updating.App.EmitAppEvent(updating.Name)

	updating.CurrentSlotIndex = -1
	for index, slot := range updating.App.GetSlots() {
		if slot.Version.ID == updating.App.ProposedVersion.ID {
			updating.CurrentSlotIndex = index + 1
		}
	}

	if updating.CurrentSlotIndex == -1 {
		updating.CurrentSlotIndex = 0
	}
	updating.TargetSlotIndex = updating.CurrentSlotIndex + updating.SlotCountNeedUpdate - 1

	updating.CurrentSlot, _ = updating.App.GetSlot(updating.CurrentSlotIndex)
	// rolling update on the first slot

	if updating.CurrentSlot != nil {
		if updating.CurrentSlotIndex == 0 {
			updating.CurrentSlot.SetWeight(0)
		}
		updating.CurrentSlot.KillTask()
	}
}

func (updating *StateUpdating) OnExit() {
	logrus.Debug("state updating OnExit")
}

func (updating *StateUpdating) Step() {
	logrus.Debug("state updating step")

	if (updating.CurrentSlot.StateIs(SLOT_STATE_REAP) ||
		updating.CurrentSlot.StateIs(SLOT_STATE_TASK_KILLED) ||
		updating.CurrentSlot.StateIs(SLOT_STATE_TASK_FINISHED) ||
		updating.CurrentSlot.Abnormal()) &&
		updating.CurrentSlotIndex <= updating.TargetSlotIndex {

		logrus.Infof("archive current task")
		updating.CurrentSlot.Archive()
		if updating.App.IsFixed() {
			updating.CurrentSlot.Ip = updating.App.ProposedVersion.IP[updating.CurrentSlotIndex]
		}
		updating.CurrentSlot.DispatchNewTask(updating.App.ProposedVersion)

	} else if updating.CurrentSlot.StateIs(SLOT_STATE_TASK_RUNNING) &&
		updating.CurrentSlot.Healthy() &&
		updating.CurrentSlotIndex < updating.TargetSlotIndex {

		updating.CurrentSlotIndex += 1
		updating.CurrentSlot, _ = updating.App.GetSlot(updating.CurrentSlotIndex)
		updating.CurrentSlot.KillTask()

	} else if updating.CurrentSlot.StateIs(SLOT_STATE_TASK_RUNNING) &&
		updating.CurrentSlot.Healthy() &&
		updating.CurrentSlotIndex == updating.TargetSlotIndex {

		if updating.CurrentSlotIndex == len(updating.App.GetSlots())-1 {
			logrus.Debug("state updating step, updating done,  all slots updated")

			updating.App.CurrentVersion = updating.App.ProposedVersion
			updating.App.ProposedVersion = nil

			updating.App.TransitTo(APP_STATE_NORMAL)
		} else {
			logrus.Debug("state updating step, updating done,  not all slots updated")
		}
	} else {
		logrus.Info("state updating step, do nothing")
	}
}

func (updating *StateUpdating) StateName() string {
	return updating.Name
}

// state machine can transit to any state if current state is updating
func (updating *StateUpdating) CanTransitTo(targetState string) bool {
	logrus.Debugf("state updating CanTransitTo %s", targetState)

	if utils.SliceContains(ValidNextTransitionState, targetState) {
		return true
	}

	return false
}
