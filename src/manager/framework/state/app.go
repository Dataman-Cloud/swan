package state

import (
	"time"

	"github.com/Dataman-Cloud/swan/src/types"
)

type AppMode string

var (
	APP_MODE_FIXED      AppMode = "fixed"
	APP_MODE_REPLICATES AppMode = "replicates"
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

	ClusterId string
}

func NewApp(version *types.Version, allocator *OfferAllocator, ClusterId string) (*App, error) {
	app := &App{
		Versions:          []*types.Version{version},
		Slots:             make([]*Slot, 0),
		CurrentVersion:    version,
		OfferAllocatorRef: allocator,
		AppId:             version.AppId,
		ClusterId:         ClusterId,

		Created: time.Now(),
		Updated: time.Now(),
	}

	for i := 0; i < int(version.Instances); i++ {
		slot := NewSlot(app, version, i)
		app.Slots = append(app.Slots, slot)
		app.OfferAllocatorRef.PutSlotBackToPendingQueue(slot)
	}

	return app, nil
}

// delete a application and all related objects: versions, tasks, slots, proxies, dns record
func (app *App) Delete() error {
	return nil
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
	return 0
}

func (app *App) RollbackInstances() int {
	return 0
}
