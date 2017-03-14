package state

import (
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
)

// load app data frm persistent data
func LoadAppData(userEventChan chan *event.UserEvent) (map[string]*App, error) {
	raftApps, err := persistentStore.ListApps()
	if err != nil {
		return nil, err
	}

	apps := make(map[string]*App)

	for _, raftApp := range raftApps {
		app := &App{
			ID:      raftApp.ID,
			Name:    raftApp.Name,
			Created: time.Unix(0, raftApp.CreatedAt),
			Updated: time.Unix(0, raftApp.UpdatedAt),
			slots:   make(map[int]*Slot),
		}

		app.UserEventChan = userEventChan

		if raftApp.Version != nil {
			app.CurrentVersion = VersionFromRaft(raftApp.Version)
			app.Mode = AppMode(raftApp.Version.Mode)
		} else {
			// TODO raftApp.Version should not be nil but we need more infomation to
			// find the reason cause the raftApp.Version nil
			logrus.Errorf("app: %s version was nil", app.ID)
		}

		if raftApp.ProposedVersion != nil {
			app.ProposedVersion = VersionFromRaft(raftApp.ProposedVersion)
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
			app.SetSlot(int(slot.Index), slot)
		}

		apps[app.ID] = app
	}

	return apps, nil
}

func LoadAppSlots(app *App) ([]*Slot, error) {
	raftSlots, err := persistentStore.ListSlots(app.ID)
	if err != nil {
		return nil, err
	}

	var slots []*Slot
	for _, raftSlot := range raftSlots {
		slot := SlotFromRaft(raftSlot)

		raftTasks, err := persistentStore.ListTasks(app.ID, slot.ID)

		if err != nil {
			return nil, err
		}

		var tasks []*Task
		for _, raftTask := range raftTasks {
			tasks = append(tasks, TaskFromRaft(raftTask))
		}
		slot.TaskHistory = tasks

		slot.CurrentTask.Slot = slot

		if slot.CurrentTask.Version == nil {
			slot.CurrentTask.Version = app.CurrentVersion
		}
		slot.App = app
		// TODO yaoyun
		slot.Version = app.CurrentVersion

		slots = append(slots, slot)
	}

	return slots, nil
}

func LoadOfferAllocatorMap() (map[string]*OfferInfo, error) {
	m := make(map[string]*OfferInfo)
	if list, err := persistentStore.ListOfferallocatorItems(); err == nil {
		for _, item := range list {
			slotId, offerInfo := OfferAllocatorItemFromRaft(item)
			m[slotId] = offerInfo
		}
	} else {
		return m, err
	}

	return m, nil
}
