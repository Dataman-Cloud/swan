package state

import (
	"github.com/Dataman-Cloud/swan/src/types"
)

type AppMode string

var (
	APP_MODE_FIXED      AppMode = "fixed"
	APP_MODE_REPLICATES AppMode = "replicates"
)

type App struct {
	Versions []*types.Version
	Slots    []*Slot

	CurrentVersion  *types.Version
	ProposedVersion *types.Version
	Mode            AppMode // fixed or repliactes
}

func NewApp(version *types.Version) (*App, error) {
	app := &App{
		Versions:       []*types.Version{version},
		Slots:          make([]*Slot, 0),
		CurrentVersion: version,
	}

	for i := 0; i < int(version.Instances); i++ {
		app.Slots = append(app.Slots, NewSlot(app, version, i))
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
