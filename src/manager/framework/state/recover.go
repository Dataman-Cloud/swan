package state

import (
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/framework/mesos_connector"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"
	"github.com/Dataman-Cloud/swan/src/types"
)

// load app data frm persistent data
func LoadAppData(allocator *OfferAllocator, mesosConnector *mesos_connector.MesosConnector,
	scontext *swancontext.SwanContext) (map[string]*App, error) {
	raftApps, err := persistentStore.ListApps()
	if err != nil {
		return nil, err
	}

	apps := make(map[string]*App)

	for _, raftApp := range raftApps {
		app := &App{
			AppId:               raftApp.ID,
			CurrentVersion:      VersionFromRaft(raftApp.Version),
			State:               raftApp.State,
			Mode:                AppMode(raftApp.Version.Mode),
			Created:             time.Unix(0, raftApp.CreatedAt),
			Updated:             time.Unix(0, raftApp.UpdatedAt),
			Scontext:            scontext,
			Slots:               make(map[int]*Slot),
			InvalidateCallbacks: make(map[string][]AppInvalidateCallbackFuncs),
			MesosConnector:      mesosConnector,
			OfferAllocatorRef:   allocator,
		}

		raftVersions, err := persistentStore.ListVersions(raftApp.ID)
		if err != nil {
			return nil, err
		}

		var versions []*types.Version
		for _, raftVersion := range raftVersions {
			versions = append(versions, VersionFromRaft(raftVersion))
		}

		app.Versions = versions

		slots, err := LoadAppSlots(app)
		if err != nil {
			return nil, err
		}

		for _, slot := range slots {
			app.Slots[int(slot.Index)] = slot
		}

		apps[app.AppId] = app
	}

	return apps, nil
}

func LoadAppSlots(app *App) ([]*Slot, error) {
	raftSlots, err := persistentStore.ListSlots(app.AppId)
	if err != nil {
		return nil, err
	}

	var slots []*Slot
	for _, raftSlot := range raftSlots {
		slot := SlotFromRaft(raftSlot)

		raftTasks, err := persistentStore.ListTasks(app.AppId, slot.Id)
		if err != nil {
			return nil, err
		}

		var tasks []*Task
		for _, raftTask := range raftTasks {
			tasks = append(tasks, TaskFromRaft(raftTask))
		}
		slot.TaskHistory = tasks

		slot.CurrentTask.Slot = slot
		slot.CurrentTask.MesosConnector = app.MesosConnector

		if slot.CurrentTask.Version == nil {
			slot.CurrentTask.Version = app.CurrentVersion
		}
		slot.App = app

		slot.StatesCallbacks = make(map[string][]SlotStateCallbackFuncs)

		slots = append(slots, slot)
	}

	return slots, nil
}
