package state

import (
	"github.com/Dataman-Cloud/swan/src/types"
)

type Task struct {
	App     *App
	Version *types.Version
	Slot    *Slot
}

func NewTask() *Task {
	task := &Task{}

	return task
}
