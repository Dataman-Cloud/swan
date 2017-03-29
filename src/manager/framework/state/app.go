package state

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	eventbus "github.com/Dataman-Cloud/swan/src/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework/connector"
	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework/store"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/utils"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

type AppMode string

var (
	APP_MODE_FIXED      AppMode = "fixed"
	APP_MODE_REPLICATES AppMode = "replicates"
)

var persistentStore store.Store

func SetStore(newStore store.Store) {
	persistentStore = newStore
}

type App struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Versions []*types.Version `json:"versions"`

	slotsLock sync.Mutex
	Slots     map[int]*Slot `json:"slots"`

	// app run with CurrentVersion config
	CurrentVersion *types.Version `json:"current_version"`
	// use when app updated, ProposedVersion can either be commit or revert
	ProposedVersion *types.Version `json:"proposed_version"`

	Mode AppMode `json:"mode"` // fixed or repliactes

	Created time.Time
	Updated time.Time

	StateMachine *StateMachine
	ClusterID    string

	UserEventChan chan *event.UserEvent
}

type AppsByUpdated []*App

func (a AppsByUpdated) Len() int           { return len(a) }
func (a AppsByUpdated) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AppsByUpdated) Less(i, j int) bool { return a[i].Updated.After(a[j].Updated) } // NOTE(xychu): Desc order

func NewApp(version *types.Version,
	userEventChan chan *event.UserEvent) (*App, error) {
	appID := fmt.Sprintf("%s-%s-%s", version.AppName, version.RunAs, connector.Instance().ClusterID)
	existingApp, _ := persistentStore.GetApp(appID)
	if existingApp != nil {
		return nil, errors.New("app already exists")
	}

	err := validateAndFormatVersion(version)
	if err != nil {
		return nil, err
	}

	app := &App{
		Versions:       []*types.Version{},
		Slots:          make(map[int]*Slot),
		CurrentVersion: version,
		ID:             appID,
		Name:           version.AppName,
		ClusterID:      connector.Instance().ClusterID,
		Created:        time.Now(),
		Updated:        time.Now(),
		UserEventChan:  userEventChan,
	}

	if version.Mode == "fixed" {
		app.Mode = APP_MODE_FIXED
	} else { // if no mode specified, default should be replicates
		app.Mode = APP_MODE_REPLICATES
	}

	version.ID = fmt.Sprintf("%d", time.Now().Unix())
	if version.AppVersion == "" {
		version.AppVersion = version.ID
	}

	app.StateMachine = NewStateMachine()
	app.StateMachine.Start(NewStateCreating(app))

	app.create()

	app.SaveVersion(app.CurrentVersion)

	return app, nil
}

// also need user pass ip here
func (app *App) ScaleUp(newInstances int, newIps []string) error {
	if !app.StateMachine.CanTransitTo(APP_STATE_SCALE_UP) {
		return errors.New(fmt.Sprintf("state machine can not transit from state: %s to state: %s",
			app.StateMachine.ReadableState(), APP_STATE_SCALE_UP))
	}

	if newInstances <= 0 {
		return errors.New("specify instances num want to increase")
	}

	if app.IsFixed() && len(newIps) != newInstances {
		return fmt.Errorf("please provide %d unique ip", newInstances)
	}

	app.CurrentVersion.IP = append(app.CurrentVersion.IP, newIps...)
	app.CurrentVersion.Instances = int32(len(app.Slots) + newInstances)
	app.Updated = time.Now()

	app.Touch()

	return app.TransitTo(APP_STATE_SCALE_UP)
}

func (app *App) ScaleDown(removeInstances int) error {
	if !app.StateMachine.CanTransitTo(APP_STATE_SCALE_DOWN) {
		return errors.New(fmt.Sprintf("state machine can not transit from state: %s to state: %s",
			app.StateMachine.ReadableState(), APP_STATE_SCALE_DOWN))
	}

	if removeInstances <= 0 {
		return errors.New("please specify at least 1 task to scale-down")
	}

	if removeInstances > len(app.Slots) {
		return fmt.Errorf("no more than %d tasks can be shutdown", app.CurrentVersion.Instances)
	}

	app.CurrentVersion.Instances = int32(len(app.Slots) - removeInstances)
	app.Updated = time.Now()

	app.Touch()

	return app.TransitTo(APP_STATE_SCALE_DOWN)
}

// delete a application and all related objects: versions, tasks, slots, proxies, dns record
func (app *App) Delete() error {
	if !app.StateMachine.CanTransitTo(APP_STATE_DELETING) {
		return errors.New(fmt.Sprintf("state machine can not transit from state: %s to state: %s",
			app.StateMachine.ReadableState(), APP_STATE_DELETING))
	}

	return app.TransitTo(APP_STATE_DELETING)
}

func (app *App) Update(version *types.Version, store store.Store) error {
	if !app.StateMachine.CanTransitTo(APP_STATE_UPDATING) || app.ProposedVersion != nil {
		return errors.New(fmt.Sprintf("state machine can not transit from state: %s to state: %s",
			app.StateMachine.ReadableState(), APP_STATE_UPDATING))
	}

	if err := validateAndFormatVersion(version); err != nil {
		return err
	}

	version.ID = fmt.Sprintf("%d", time.Now().Unix())
	if version.AppVersion == "" {
		version.AppVersion = version.ID
	}

	if err := app.checkProposedVersionValid(version); err != nil {
		return err
	}

	if app.CurrentVersion == nil {
		return errors.New("update failed: current version was losted")
	}

	app.ProposedVersion = version

	app.SaveVersion(app.ProposedVersion)
	app.Touch()

	return app.TransitTo(APP_STATE_UPDATING, 1)
}

func (app *App) ProceedingRollingUpdate(instances int) error {
	if !app.StateMachine.CanTransitTo(APP_STATE_UPDATING) || app.ProposedVersion == nil {
		return errors.New(fmt.Sprintf("state machine can not transit from state: %s to state: %s",
			app.StateMachine.ReadableState(), APP_STATE_UPDATING))
	}

	updatedCount := 0
	for index, slot := range app.GetSlots() {
		if slot.Version.ID == app.ProposedVersion.ID {
			updatedCount = index + 1
		}
	}

	// when instances is -1, indicates all remaining
	if instances == -1 {
		instances = len(app.Slots) - updatedCount
	}

	if instances < 1 {
		return errors.New("please specify how many instance want proceeding the update")
	}

	if updatedCount+instances > len(app.GetSlots()) {
		return errors.New(fmt.Sprintf("only %d tasks left need to be updated now", len(app.GetSlots())-updatedCount))
	}

	return app.TransitTo(APP_STATE_UPDATING, instances)
}

func (app *App) CancelUpdate() error {
	if !app.StateMachine.CanTransitTo(APP_STATE_CANCEL_UPDATE) || app.ProposedVersion == nil {
		return errors.New(fmt.Sprintf("state machine can not transit from state: %s to state: %s",
			app.StateMachine.ReadableState(), APP_STATE_CANCEL_UPDATE))
	}

	if app.CurrentVersion == nil {
		return errors.New("cancel update failed: current version was nil")
	}

	return app.TransitTo(APP_STATE_CANCEL_UPDATE)
}

func (app *App) ServiceDiscoveryURL() string {
	return strings.ToLower(strings.Replace(app.ID, "-", ".", -1))
}

func (app *App) IsReplicates() bool {
	return app.Mode == APP_MODE_REPLICATES
}

func (app *App) IsFixed() bool {
	return app.Mode == APP_MODE_FIXED
}

func (app *App) EmitAppEvent(stateString string) {
	eventType := ""
	switch stateString {
	case APP_STATE_CREATING:
		eventType = eventbus.EventTypeAppStateCreating
	case APP_STATE_DELETING:
		eventType = eventbus.EventTypeAppStateDeletion
	case APP_STATE_NORMAL:
		eventType = eventbus.EventTypeAppStateNormal
	case APP_STATE_UPDATING:
		eventType = eventbus.EventTypeAppStateUpdating
	case APP_STATE_CANCEL_UPDATE:
		eventType = eventbus.EventTypeAppStateCancelUpdate
	case APP_STATE_SCALE_UP:
		eventType = eventbus.EventTypeAppStateScaleUp
	case APP_STATE_SCALE_DOWN:
		eventType = eventbus.EventTypeAppStateScaleDown
	default:
	}

	e := &eventbus.Event{Type: eventType}
	e.AppID = app.ID
	e.Payload = &types.AppInfoEvent{
		AppID:     app.ID,
		Name:      app.Name,
		ClusterID: app.ClusterID,
		RunAs:     app.CurrentVersion.RunAs,
	}

	eventbus.WriteEvent(e)
}

func (app *App) StateIs(state string) bool {
	return app.StateMachine.Is(state)
}

func (app *App) TransitTo(targetState string, args ...interface{}) error {
	err := app.StateMachine.TransitTo(app.stateFactory(targetState, args...))
	if err != nil {
		return err
	}

	app.Touch()

	return nil
}

func (app *App) RemoveSlot(index int) {
	if slot, found := app.GetSlot(index); found {
		OfferAllocatorInstance().RemoveSlotFromAllocator(slot)
		slot.Remove()

		app.slotsLock.Lock()
		delete(app.Slots, index)
		app.slotsLock.Unlock()

		app.Touch()
	}
}

func (app *App) stateFactory(stateName string, args ...interface{}) State {
	switch stateName {
	case APP_STATE_NORMAL:
		return NewStateNormal(app)
	case APP_STATE_CREATING:
		return NewStateCreating(app)
	case APP_STATE_DELETING:
		return NewStateDeleting(app)
	case APP_STATE_SCALE_UP:
		return NewStateScaleUp(app)
	case APP_STATE_SCALE_DOWN:
		return NewStateScaleDown(app)
	case APP_STATE_UPDATING:
		slotCountNeedUpdate, ok := args[0].(int)
		if !ok {
			slotCountNeedUpdate = 1
		}
		return NewStateUpdating(app, slotCountNeedUpdate)

	case APP_STATE_CANCEL_UPDATE:
		return NewStateCancelUpdate(app)
	default:
		panic(errors.New("unrecognized state"))
	}
}

func (app *App) GetSlots() []*Slot {
	slots := make([]*Slot, 0)
	for _, v := range app.Slots {
		slots = append(slots, v)
	}

	slotsById := SlotsById(slots)
	sort.Sort(slotsById)

	return slotsById
}

func (app *App) GetSlot(index int) (*Slot, bool) {
	slot, ok := app.Slots[index]
	return slot, ok
}

func (app *App) SetSlot(index int, slot *Slot) {
	app.slotsLock.Lock()
	app.Slots[index] = slot
	app.slotsLock.Unlock()

	app.Touch()
}

func (app *App) Step() {
	app.StateMachine.Step()
	app.Touch()
}

// make sure proposed version is valid then applied it to field ProposedVersion
func (app *App) checkProposedVersionValid(version *types.Version) error {
	// mode can not change
	if version.Mode != app.CurrentVersion.Mode {
		return fmt.Errorf("mode can not change when update app, current version is %s", app.CurrentVersion.Mode)
	}
	// runAs can not change
	if version.RunAs != app.CurrentVersion.RunAs {
		return fmt.Errorf("runAs can not change when update app, current version is %s", app.CurrentVersion.RunAs)
	}

	// appVersion should not equal
	if version.AppVersion == app.CurrentVersion.AppVersion {
		return fmt.Errorf("app version %s exists, choose another one", version.AppVersion)
	}
	// app instances should same as current instances
	if version.Instances != app.CurrentVersion.Instances {
		return fmt.Errorf("instances can not change when update app, current version is %d", app.CurrentVersion.Instances)
	}
	// fixed app IP length should be same as current instances
	if app.IsFixed() && int32(len(version.IP)) != app.CurrentVersion.Instances {
		return fmt.Errorf("Fixed mode App IP length can not change when update app, current version is %d", app.CurrentVersion.Instances)
	}
	return nil
}

func validateAndFormatVersion(version *types.Version) error {
	if version.Container == nil {
		return errors.New("swan only support mesos docker containerization, no container found")
	}

	if version.Container.Docker == nil {
		return errors.New("swan only support mesos docker containerization, no container found")
	}

	if version.AppName == "" {
		return errors.New("invalid appName: appName was empty")
	}

	if version.Instances == 0 {
		return errors.New("invalid instances: instances must be specified and should greater than 0")
	}

	version.AppName = strings.TrimSpace(version.AppName)

	r, _ := regexp.Compile("([A-Z]+)|([\\-\\.\\$\\*\\+\\?\\{\\}\\(\\)\\[\\]\\|]+)")
	errMsg := errors.New(`must be lower case characters and should not contain following special characters "-.$*?{}()[]|"`)

	//validation of AppId
	match := r.MatchString(version.AppName)
	if match {
		return fmt.Errorf("invalid app id [%s]: %s", version.AppName, errMsg)
	}

	//validation of RunAs
	match = r.MatchString(version.RunAs)
	if match {
		return fmt.Errorf("invalid runAs [%s]: %s", version.RunAs, errMsg)
	}

	match = r.MatchString(version.Container.Docker.Network)
	if match {
		return fmt.Errorf("invalid network [%s]: %s", version.Container.Docker.Network, errMsg)
	}

	if len(version.RunAs) == 0 {
		return errors.New("runAs should not empty")
	}

	if len(version.Mode) == 0 {
		version.Mode = string(APP_MODE_REPLICATES)
	}

	if (version.Mode != string(APP_MODE_REPLICATES)) && (version.Mode != string(APP_MODE_FIXED)) {
		return fmt.Errorf("enrecognized app mode %s", version.Mode)
	}

	// validation for fixed mode application
	if version.Mode == string(APP_MODE_FIXED) {
		if len(version.IP) != int(version.Instances) {
			return fmt.Errorf("should provide exactly %d ip for FIXED type app", version.Instances)
		}

		if len(version.Container.Docker.PortMappings) > 0 {
			return errors.New("fixed mode application doesn't support portmapping")
		}
	}

	// validation for replicates mode app
	if version.Mode == string(APP_MODE_REPLICATES) {
		// the only network driver should be **bridge**
		if !utils.SliceContains([]string{"bridge", "host"}, strings.ToLower(version.Container.Docker.Network)) {
			return errors.New("replicates mode app suppose the only network driver should be bridge or host")
		}

		// portMapping.Name should be mandatory
		for _, portmapping := range version.Container.Docker.PortMappings {
			if strings.TrimSpace(portmapping.Name) == "" {
				return errors.New("each port mapping should have a uniquely identified name")
			}
		}

		if strings.ToLower(version.Container.Docker.Network) == "host" {
			// portMapping.Name should be mandatory
			for _, portmapping := range version.Container.Docker.PortMappings {
				if portmapping.ContainerPort != 0 {
					return errors.New("containerPort not recongnizable for docker host network, port is mandatory")
				}
			}
		}

		portNames := make([]string, 0)
		for _, portmapping := range version.Container.Docker.PortMappings {
			portNames = append(portNames, portmapping.Name)
		}

		// portName should be unique
		if !utils.SliceUnique(portNames) {
			return errors.New("each port mapping should have a uniquely identified name")
		}

		// portName for health check should mandatory
		if version.HealthCheck != nil {
			protocol, portName := version.HealthCheck.Protocol, version.HealthCheck.PortName
			// portName should present in dockers' portMappings definition
			// portName if manditory for non-cmd health checks
			if strings.ToLower(protocol) != "cmd" && !utils.SliceContains(portNames, portName) {
				return fmt.Errorf("portname in healthCheck section should match that defined in portMappings")
			}

			if !utils.SliceContains([]string{"tcp", "http", "TCP", "HTTP", "cmd", "CMD"}, protocol) {
				return fmt.Errorf("doesn't recoginized protocol %s for health check", protocol)
			}

			if strings.ToLower(protocol) == "http" {
				if len(version.HealthCheck.Path) == 0 {
					return fmt.Errorf("no path provided for health check with %s protocol", protocol)
				}
			}

			if strings.ToLower(protocol) == "cmd" {
				if len(version.HealthCheck.Value) == 0 {
					return fmt.Errorf("no value provided for health check with %s protocol", protocol)
				}
			}
		}
	} else {
		if version.HealthCheck != nil {
			protocol := version.HealthCheck.Protocol
			if !utils.SliceContains([]string{"cmd", "CMD"}, protocol) {
				return fmt.Errorf("doesn't recoginized protocol %s for health check for fixed type app", protocol)
			}

			if len(version.HealthCheck.Value) == 0 {
				return fmt.Errorf("no value provided for health check with %s", protocol)
			}
		}
	}

	// validate constraints are all valid
	if len(version.Constraints) > 0 {
		evalStatement, err := ParseConstraint(strings.ToLower(version.Constraints))
		if err != nil {
			return err
		}

		err = evalStatement.Valid()
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *App) SaveVersion(version *types.Version) {
	app.Versions = append(app.Versions, version)
	WithConvertVersion(context.TODO(), app.ID, version, nil, persistentStore.CreateVersion)
}

// 1, remove app from persisted storage
// 2, other cleanup process
func (app *App) Remove() {
	app.remove()
}

// storage related
func (app *App) Touch() {
	app.update()
}

func (app *App) update() {
	logrus.Debugf("update app %s", app.ID)
	WithConvertApp(context.TODO(), app, nil, persistentStore.UpdateApp)
}

func (app *App) create() {
	logrus.Debugf("create app %s", app.ID)
	WithConvertApp(context.TODO(), app, nil, persistentStore.CreateApp)
}

func (app *App) remove() {
	logrus.Debugf("remove app %s", app.ID)
	persistentStore.DeleteApp(context.TODO(), app.ID, nil)
}
