package state

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	swanevent "github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework/mesos_connector"
	"github.com/Dataman-Cloud/swan/src/manager/framework/store"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
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

var persistentStore store.Store

func SetStore(newStore store.Store) {
	persistentStore = newStore
}

type App struct {
	// app name
	AppId    string           `json:"appId"`
	Versions []*types.Version `json:"versions"`

	slotsLock sync.Mutex
	slots     map[int]*Slot `json:"slots"`

	Scontext *swancontext.SwanContext

	// app run with CurrentVersion config
	CurrentVersion *types.Version `json:"current_version"`
	// use when app updated, ProposedVersion can either be commit or revert
	ProposedVersion *types.Version `json:"proposed_version"`

	Mode AppMode `json:"mode"` // fixed or repliactes

	OfferAllocatorRef *OfferAllocator
	Created           time.Time
	Updated           time.Time

	State string

	MesosConnector *mesos_connector.MesosConnector
}

func NewApp(version *types.Version,
	allocator *OfferAllocator,
	MesosConnector *mesos_connector.MesosConnector,
	scontext *swancontext.SwanContext) (*App, error) {
	err := validateAndFormatVersion(version)
	if err != nil {
		return nil, err
	}

	app := &App{
		Versions:          []*types.Version{version},
		slots:             make(map[int]*Slot),
		CurrentVersion:    version,
		OfferAllocatorRef: allocator,
		AppId:             version.AppId,
		MesosConnector:    MesosConnector,
		Scontext:          scontext,
		Created:           time.Now(),
		Updated:           time.Now(),
	}

	if version.Mode == "fixed" {
		app.Mode = APP_MODE_FIXED
	} else { // if no mode specified, default should be replicates
		app.Mode = APP_MODE_REPLICATES
	}
	version.ID = fmt.Sprintf("%d", time.Now().Unix())

	if err := WithConvertApp(context.TODO(), app, nil, persistentStore.CreateApp); err != nil {
		return nil, err
	}

	for i := 0; i < int(version.Instances); i++ {
		slot := NewSlot(app, version, i)
		//if err := WithConvertSlot(context.TODO(), slot, func() { app.SetSlot(slot.Index, slot) }, persistentStore.CreateSlot); err != nil {
		//return nil, err
		//}

		app.SetSlot(i, slot)

	}

	app.SetState(APP_STATE_MARK_FOR_CREATING)

	return app, nil
}

// also need user pass ip here
func (app *App) ScaleUp(newInstances int, newIps []string) error {
	if !app.StateIs(APP_STATE_NORMAL) {
		return errors.New("app not in normal state")
	}

	if newInstances <= 0 {
		return errors.New("specify instances num want to increase")
	}

	if app.IsFixed() && len(newIps) != newInstances {
		return errors.New(fmt.Sprintf("please provide %d unique ip", newInstances))
	}

	app.CurrentVersion.Ip = append(app.CurrentVersion.Ip, newIps...)
	app.CurrentVersion.Instances += int32(newInstances)
	app.Updated = time.Now()

	app.SetState(APP_STATE_MARK_FOR_SCALE_UP)

	for i := newInstances; i > 0; i-- {
		slotIndex := int(app.CurrentVersion.Instances) - i
		slot := NewSlot(app, app.CurrentVersion, slotIndex)
		app.SetSlot(slotIndex, slot)
	}
	return nil
}

func (app *App) ScaleDown(removeInstances int) error {
	if !app.StateIs(APP_STATE_NORMAL) {
		return errors.New("app not in normal state")
	}

	if removeInstances >= int(app.CurrentVersion.Instances) {
		return errors.New(fmt.Sprintf("no more than %d instances can be shutdown", app.CurrentVersion.Instances))
	}

	app.CurrentVersion.Instances = int32(int(app.CurrentVersion.Instances) - removeInstances)
	app.Updated = time.Now()

	app.SetState(APP_STATE_MARK_FOR_SCALE_DOWN)

	for i := removeInstances; i > 0; i-- {
		slotIndex := int(app.CurrentVersion.Instances) + i - 1
		defer func(slotIndex int) {
			slot := app.slots[slotIndex]
			slot.Kill()
		}(slotIndex)
	}

	return nil
}

// delete a application and all related objects: versions, tasks, slots, proxies, dns record
func (app *App) Delete() error {
	app.SetState(APP_STATE_MARK_FOR_DELETION)

	for _, slot := range app.slots {
		slot.Kill()
	}

	return nil
}

// update application by follower steps
// 1. check app state: if app state if not APP_STATE_NORMAL or app's propose version is not nil
//    we can not update app, because that means target app maybe is in updateing.
// 2. set the new version to the app's propose version.
// 3. persist app data, and set the app's state to APP_STATE_MARK_FOR_UPDATING
// 4. update slot version to propose version
// 5. after all slot version update success. put the current version to version history and set the
//    propose version as current version, set propose version to nil.
// 6. set app's state to APP_STATE_NORMAL.
func (app *App) Update(version *types.Version, store store.Store) error {
	if !app.StateIs(APP_STATE_NORMAL) || app.ProposedVersion != nil {
		return errors.New("app not in normal state")
	}

	if err := app.checkProposedVersionValid(version); err != nil {
		return err
	}

	if app.CurrentVersion == nil {
		return errors.New("update failed: current version was losted")
	}

	app.SetState(APP_STATE_MARK_FOR_UPDATING)

	version.ID = fmt.Sprintf("%d", time.Now().Unix())
	version.PerviousVersionID = app.CurrentVersion.ID
	app.ProposedVersion = version

	for i := 0; i < 1; i++ { // current we make first slot update
		slot := app.slots[i]
		slot.Update(app.ProposedVersion, true)
	}

	return nil
}

func (app *App) ProceedingRollingUpdate(instances int) error {
	if app.ProposedVersion == nil {
		return errors.New("app not in rolling update state")
	}

	if instances < 1 {
		return errors.New("please specify how many instance want proceeding the update")
	}

	if (instances + app.RollingUpdateInstances()) > int(app.CurrentVersion.Instances) {
		return errors.New("update instances count exceed the maximum instances number")
	}

	for i := 0; i < instances; i++ {
		slotIndex := i + app.RollingUpdateInstances()
		defer func(slotIndex int) { // RollingUpdateInstances() has side effects in the loop
			slot := app.slots[slotIndex]
			slot.Update(app.ProposedVersion, true)
		}(slotIndex)
	}

	return nil
}

func (app *App) CancelUpdate() error {
	if app.State != APP_STATE_MARK_FOR_UPDATING || app.ProposedVersion == nil {
		return errors.New("app not in updating state")
	}

	if app.CurrentVersion == nil {
		return errors.New("cancel update failed: current version was nil")
	}

	for i := app.RollingUpdateInstances() - 1; i >= 0; i-- {
		slot := app.slots[i]
		slot.Update(app.CurrentVersion, false)
	}

	return nil
}

func (app *App) IsReplicates() bool {
	return app.Mode == APP_MODE_REPLICATES
}

func (app *App) IsFixed() bool {
	return app.Mode == APP_MODE_FIXED
}

func (app *App) SetState(state string) {
	app.State = state
	logrus.Infof("app %s now has state %s", app.AppId, app.State)
}

func (app *App) StateIs(state string) bool {
	return app.State == state
}

func (app *App) RunningInstances() int {
	runningInstances := 0
	for _, slot := range app.slots {
		if slot.StateIs(SLOT_STATE_TASK_RUNNING) {
			runningInstances += 1
		}
	}

	return runningInstances
}

func (app *App) RollingUpdateInstances() int {
	rollingUpdateInstances := 0
	for _, slot := range app.slots {
		if slot.MarkForRollingUpdate {
			rollingUpdateInstances += 1
		}
	}

	return rollingUpdateInstances
}

func (app *App) MarkForDeletionInstances() int {
	markForDeletionInstances := 0
	for _, slot := range app.slots {
		if slot.MarkForDeletion {
			markForDeletionInstances += 1
		}
	}

	return markForDeletionInstances
}

func (app *App) CanBeCleanAfterDeletion() bool {
	return app.StateIs(APP_STATE_MARK_FOR_DELETION) && len(app.slots) == 0
}

func (app *App) RemoveSlot(index int) {
	app.slotsLock.Lock()
	defer app.slotsLock.Unlock()

	delete(app.slots, index)
}

func (app *App) GetSlot(index int) (*Slot, bool) {
	slot, ok := app.slots[index]
	return slot, ok
}

func (app *App) GetSlots() []*Slot {
	slots := make([]*Slot, 0)
	for _, v := range app.slots {
		slots = append(slots, v)
	}

	slotsById := SlotsById(slots)
	sort.Sort(slotsById)

	return slotsById
}

func (app *App) SetSlot(index int, slot *Slot) {
	app.slotsLock.Lock()
	defer app.slotsLock.Unlock()

	app.slots[index] = slot
}

func (app *App) Reevaluate() {
	switch app.State {
	case APP_STATE_NORMAL:
	case APP_STATE_MARK_FOR_DELETION:
	case APP_STATE_MARK_FOR_UPDATING:
		// when updating done
		if (app.RollingUpdateInstances() == int(app.CurrentVersion.Instances)) &&
			(app.RunningInstances() == int(app.CurrentVersion.Instances)) { // not perfect as when instances number increase, all instances running might be hard to acheive
			app.SetState(APP_STATE_NORMAL)

			app.CurrentVersion = app.ProposedVersion
			app.Versions = append(app.Versions, app.CurrentVersion)
			app.ProposedVersion = nil

			for _, slot := range app.slots {
				slot.MarkForRollingUpdate = false
			}
		}

		// when update cancelled
		if app.slots[0].Version == app.CurrentVersion && // until the first slot has updated to CurrentVersion
			app.RunningInstances() == int(app.CurrentVersion.Instances) { // not perfect as when instances number increase, all instances running might be hard to achieve
			app.SetState(APP_STATE_NORMAL)
			app.ProposedVersion = nil
		}

	case APP_STATE_MARK_FOR_CREATING:
		if app.RunningInstances() == int(app.CurrentVersion.Instances) {
			app.SetState(APP_STATE_NORMAL)
		}

	case APP_STATE_MARK_FOR_SCALE_UP:
		if app.StateIs(APP_STATE_MARK_FOR_SCALE_UP) && (app.RunningInstances() == int(app.CurrentVersion.Instances)) {
			app.SetState(APP_STATE_NORMAL)
		}

	case APP_STATE_MARK_FOR_SCALE_DOWN:
		if len(app.slots) == int(app.CurrentVersion.Instances) &&
			app.MarkForDeletionInstances() == 0 {
			app.SetState(APP_STATE_NORMAL)
		}

	default:
	}
}

func (app *App) EmitEvent(swanEvent *swanevent.Event) {
	app.Scontext.EventBus.EventChan <- swanEvent
}

// make sure proposed version is valid then applied it to field ProposedVersion
func (app *App) checkProposedVersionValid(version *types.Version) error {
	// mode can not changed
	// runAs can not changed
	// app instances should same as current instances
	return nil
}

func validateAndFormatVersion(version *types.Version) error {
	if len(version.Mode) == 0 {
		version.Mode = string(APP_MODE_REPLICATES)
	}

	if (version.Mode != string(APP_MODE_REPLICATES)) && (version.Mode != string(APP_MODE_FIXED)) {
		return errors.New(fmt.Sprintf("enrecognized app mode %s", version.Mode))
	}

	if version.Mode == string(APP_MODE_FIXED) {
		if len(version.Ip) != int(version.Instances) {
			return errors.New(fmt.Sprintf("should provide exactly %d ip for FIXED type app", version.Instances))
		}
	}

	return nil
}
