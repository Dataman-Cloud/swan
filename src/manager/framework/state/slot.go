package state

import (
	"github.com/Dataman-Cloud/swan/src/types"
)

const (
	SLOT_STATE_PENDING_OFFER     = "pending_offer"
	SLOT_STATE_TASK_DISPATCHED   = "task_dispatched"
	SLOT_STATE_TASK_RUNING       = "task_running"
	SLOT_STATE_TASK_FAILED       = "task_failed"
	SLOT_STATE_TASK_RESCHEDULING = "task_reschuduling"
	SLOT_STATE_TASK_STOPPING     = "task_stopping"
	SLOT_STATE_TASK_STOPPED      = "task_stopped"
)

type Slot struct {
	Id      int
	App     *App
	Version *types.Version
	State   string

	RunningTask *Task
	Tasks       []*Task
}

func NewSlot(app *App, version *types.Version, index int) *Slot {
	slot := &Slot{
		Id:      index,
		App:     app,
		Version: version,
		Tasks:   make([]*Task, 0),
		State:   SLOT_STATE_PENDING_OFFER,
	}

	return slot
}
