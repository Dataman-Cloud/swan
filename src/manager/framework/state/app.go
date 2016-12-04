package state

import (
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
	APP_STATE_NORMAL            = "normal"
	APP_STATE_SCLAEING          = "sclaling"
	APP_STATE_MARK_FOR_DELETION = "deleting"
	APP_STATE_MARK_FOR_UPDATING = "updating"
)

type App struct {
	// app name
	AppId    string           `json:"appId"`
	Versions []*types.Version `json:"versions"`
	Slots    []*Slot          `json:"slots"`

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
		Slots:             make([]*Slot, 0),
		CurrentVersion:    version,
		OfferAllocatorRef: allocator,
		AppId:             version.AppId,
		Scheduler:         Scheduler,

		Created: time.Now(),
		Updated: time.Now(),

		State: APP_STATE_NORMAL,
	}

	version.ID = fmt.Sprintf("%d", time.Now().Unix())

	for i := 0; i < int(version.Instances); i++ {
		slot := NewSlot(app, version, i)
		app.Slots = append(app.Slots, slot)
		app.OfferAllocatorRef.PutSlotBackToPendingQueue(slot)
	}

	return app, nil
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

// scale a application, both up and down
// provide enough ip if mode is fixed
func (app *App) Scale(version *types.Version) error {
	err := app.checkProposedVersionValid(version)
	if err != nil {
		panic(err)
	}

	app.ProposedVersion = version
	// succedding operations
	return nil
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
	killedSlotCount := 0
	for _, slot := range app.Slots {
		if slot.StateIs(SLOT_STATE_TASK_KILLED) || slot.StateIs(SLOT_STATE_TASK_FINISHED) || slot.StateIs(SLOT_STATE_TASK_FAILED) {
			killedSlotCount += 1
		}
	}

	return app.StateIs(APP_STATE_MARK_FOR_DELETION) && killedSlotCount == int(app.CurrentVersion.Instances)
}

// TODO
func (app *App) PersistToStorage() error {
	return nil
}
