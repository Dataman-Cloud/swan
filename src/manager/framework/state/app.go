package state

import (
	"errors"
	"fmt"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
)

type AppMode string

var (
	APP_MODE_FIXED      AppMode = "fixed"
	APP_MODE_REPLICATES AppMode = "replicates"
)

const (
	APP_STATE_NORMAL              = "normal"
	APP_STATE_MARK_FOR_CREATING   = "creating"
	APP_STATE_MARK_FOR_DELETION   = "deleting"
	APP_STATE_MARK_FOR_UPDATING   = "updating"
	APP_STATE_MARK_FOR_SCALE_UP   = "scale_up"
	APP_STATE_MARK_FOR_SCALE_DOWN = "scale_down"
)

type App struct {
	// app name
	AppId    string           `json:"appId"`
	Versions []*types.Version `json:"versions"`
	Slots    map[int]*Slot    `json:"slots"`

	// app run with CurrentVersion config
	CurrentVersion *types.Version `json:"current_version"`
	// use when app updated, ProposedVersion can either be commit or revert
	ProposedVersion *types.Version `json:"proposed_version"`

	Mode AppMode `json:"mode"` // fixed or repliactes

	OfferAllocatorRef *OfferAllocator
	Created           time.Time
	Updated           time.Time

	State string

	Scheduler *scheduler.Scheduler
}

func NewApp(version *types.Version, allocator *OfferAllocator, Scheduler *scheduler.Scheduler) (*App, error) {
	app := &App{
		Versions:          []*types.Version{version},
		Slots:             make(map[int]*Slot),
		CurrentVersion:    version,
		OfferAllocatorRef: allocator,
		AppId:             version.AppId,
		Scheduler:         Scheduler,

		Created: time.Now(),
		Updated: time.Now(),

		State: APP_STATE_MARK_FOR_CREATING,
	}

	version.ID = fmt.Sprintf("%d", time.Now().Unix())

	for i := 0; i < int(version.Instances); i++ {
		slot := NewSlot(app, version, i)
		app.Slots[i] = slot
		app.OfferAllocatorRef.PutSlotBackToPendingQueue(slot)
	}

	return app, nil
}

// also need user pass ip here
func (app *App) ScaleUp(to int) error {
	app.SetState(APP_STATE_MARK_FOR_SCALE_UP)
	if to <= int(app.CurrentVersion.Instances) {
		return errors.New("scale up expected instances size no less than current size")
	}

	for i := int(app.CurrentVersion.Instances); i < to; i++ {
		slot := NewSlot(app, app.CurrentVersion, i)
		app.Slots[i] = slot
		app.OfferAllocatorRef.PutSlotBackToPendingQueue(slot)
	}
	app.CurrentVersion.Instances = int32(to)
	app.Updated = time.Now()

	return nil
}

func (app *App) ScaleDown(to int) error {
	app.SetState(APP_STATE_MARK_FOR_SCALE_DOWN)
	if to >= int(app.CurrentVersion.Instances) {
		return errors.New("scale down expected instances size no bigger than current size")
	}

	for i := int(app.CurrentVersion.Instances); i > to; i-- {
		slot := app.Slots[i-1]
		slot.SetState(SLOT_STATE_PENDING_KILL)
	}

	app.CurrentVersion.Instances = int32(to)
	app.Updated = time.Now()

	return nil
}

// delete a application and all related objects: versions, tasks, slots, proxies, dns record
func (app *App) Delete() error {
	app.SetState(APP_STATE_MARK_FOR_DELETION)
	for _, slot := range app.Slots {
		slot.SetState(SLOT_STATE_PENDING_KILL)
	}

	return nil
}

func (app *App) SetState(state string) {
	app.State = state
	logrus.Infof("app %s now has state %s", app.AppId, app.State)
}

func (app *App) StateIs(state string) bool {
	return app.State == state
}

func (app *App) Update(version *types.Version) error {
	err := app.checkProposedVersionValid(version)
	if err != nil {
		panic(err)
	}

	app.ProposedVersion = version
	// succedding operations
	return nil
}

// make sure proposed version is valid then applied it to field ProposedVersion
func (app *App) checkProposedVersionValid(version *types.Version) error {
	return nil
}

func (app *App) RunningInstances() int {
	runningInstances := 0
	for _, slot := range app.Slots {
		if slot.StateIs(SLOT_STATE_TASK_RUNNING) {
			runningInstances += 1
		}
	}

	return runningInstances
}

func (app *App) RollbackInstances() int {
	return 0
}

func (app *App) CanBeCleanAfterDeletion() bool {
	return app.StateIs(APP_STATE_MARK_FOR_DELETION) && len(app.Slots) == 0
}

// TODO
func (app *App) PersistToStorage() error {
	return nil
}

func (app *App) InvalidateSlots() {
	switch app.State {
	case APP_STATE_MARK_FOR_DELETION:
	case APP_STATE_MARK_FOR_CREATING:
		if app.RunningInstances() == int(app.CurrentVersion.Instances) {
			app.SetState(APP_STATE_NORMAL)
		}

	case APP_STATE_MARK_FOR_SCALE_UP:
		if app.RunningInstances() == int(app.CurrentVersion.Instances) {
			app.SetState(APP_STATE_NORMAL)
		}

	case APP_STATE_MARK_FOR_SCALE_DOWN:
		if app.RunningInstances() == int(app.CurrentVersion.Instances) {
			app.SetState(APP_STATE_NORMAL)
		}

	default:

	}

	for k, slot := range app.Slots {
		if slot.MarkForDeletion && (slot.StateIs(SLOT_STATE_TASK_KILLED) || slot.StateIs(SLOT_STATE_TASK_FINISHED) || slot.StateIs(SLOT_STATE_TASK_FAILED)) {
			// TODO remove slot from OfferAllocator
			delete(app.Slots, k)
			break
		}
	}
}
