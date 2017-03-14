package state

import (
	rafttypes "github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"golang.org/x/net/context"
)

func WithConvertApp(ctx context.Context, app *App, cb func(), action func(ctx context.Context, app *rafttypes.Application, cb func()) error) error {
	raftApp := AppToRaft(app)

	return action(ctx, raftApp, cb)
}

func WithConvertSlot(ctx context.Context, slot *Slot, cb func(), action func(ctx context.Context, slot *rafttypes.Slot, cb func()) error) error {
	raftSlot := SlotToRaft(slot)

	return action(ctx, raftSlot, cb)
}

func WithConvertTask(ctx context.Context, task *Task, cb func(), action func(ctx context.Context, task *rafttypes.Task, cb func()) error) error {
	raftTask := TaskToRaft(task)

	return action(ctx, raftTask, cb)
}
