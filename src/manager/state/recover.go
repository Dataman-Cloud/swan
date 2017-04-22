package state

import (
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/store"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Sirupsen/logrus"
)

// load app data frm persistent data
func LoadAppData(userEventChan chan *event.UserEvent) map[string]*App {
	raftApps := store.DB().ListApps()

	apps := make(map[string]*App)

	for _, raftApp := range raftApps {
		app := &App{
			ID:      raftApp.ID,
			Name:    raftApp.Name,
			Created: time.Unix(0, raftApp.CreatedAt),
			Updated: time.Unix(0, raftApp.UpdatedAt),
			Slots:   make(map[int]*Slot),
		}

		app.UserEventChan = userEventChan

		if raftApp.Version != nil {
			app.CurrentVersion = VersionFromRaft(raftApp.Version)
		} else {
			// TODO raftApp.Version should not be nil but we need more infomation to
			// find the reason cause the raftApp.Version nil
			logrus.Errorf("app: %s version was nil", app.ID)
		}

		app.Mode = AppMode(raftApp.Version.Mode)

		if raftApp.ProposedVersion != nil {
			app.ProposedVersion = VersionFromRaft(raftApp.ProposedVersion)
		}

		raftVersions := store.DB().ListVersions(raftApp.ID)

		var versions []*types.Version
		for _, raftVersion := range raftVersions {
			versions = append(versions, VersionFromRaft(raftVersion))
		}

		app.Versions = versions

		slots := LoadAppSlots(app)
		for _, slot := range slots {
			app.Slots[int(slot.Index)] = slot
		}

		if raftApp.StateMachine != nil {
			app.StateMachine = StateMachineFromRaft(app, raftApp.StateMachine)
		}

		apps[app.ID] = app
	}

	return apps
}

func LoadAppSlots(app *App) []*Slot {
	raftSlots := store.DB().ListSlots(app.ID)
	var slots []*Slot
	for _, raftSlot := range raftSlots {
		slot := SlotFromRaft(raftSlot, app)

		raftTasks := store.DB().ListTaskHistory(app.ID, slot.ID)
		var tasks []*Task
		for _, raftTask := range raftTasks {
			tasks = append(tasks, TaskFromRaft(raftTask, app))
		}
		slot.TaskHistory = tasks

		slot.App = app

		slots = append(slots, slot)
	}

	return slots
}

func LoadOfferAllocatorMap() (map[string]*OfferInfo, error) {
	m := make(map[string]*OfferInfo)
	list := store.DB().ListOfferallocatorItems()
	for _, item := range list {
		slotId, offerInfo := OfferAllocatorItemFromRaft(item)
		m[slotId] = offerInfo
	}

	return m, nil
}
