package store

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

type DummyStore struct {
	Apps           map[string]*appHolder
	OfferAllocator map[string]*OfferAllocatorItem
	FrameworkId    string

	mu sync.Mutex
}

func NewDummyStore() *DummyStore {
	return &DummyStore{}
}

func (dummy *DummyStore) CreateApp(app *Application) error {
	logrus.Debug("CreateApp from DummyStore")
	return nil
}

func (dummy *DummyStore) UpdateApp(app *Application) error {
	logrus.Debug("CreateApp from DummyStore")
	return nil
}

func (dummy *DummyStore) GetApp(appId string) *Application {
	logrus.Debug("GetApp from DummyStore")
	return nil
}

func (dummy *DummyStore) ListApps() []*Application {
	logrus.Debug("ListApps from DummyStore")
	return nil
}

func (dummy *DummyStore) DeleteApp(appId string) error {
	logrus.Debug("DeleteApp from DummyStore")
	return nil
}

func (dummy *DummyStore) CreateVersion(appId string, version *Version) error {
	logrus.Debug("CreateVersion from DummyStore")
	return nil
}

func (dummy *DummyStore) GetVersion(appId, versionId string) *Version {
	logrus.Debug("GetVersion from DummyStore")
	return nil
}

func (dummy *DummyStore) ListVersions(appId string) []*Version {
	logrus.Debug("ListVersions from DummyStore")
	return nil
}

func (dummy *DummyStore) CreateSlot(slot *Slot) error {
	logrus.Debug("ListVersions from DummyStore")
	return nil
}

func (dummy *DummyStore) GetSlot(appId, slotId string) *Slot {
	logrus.Debug("ListVersions from DummyStore")
	return nil
}
func (dummy *DummyStore) ListSlots(appId string) []*Slot {
	logrus.Debug("ListVersions from DummyStore")
	return nil
}
func (dummy *DummyStore) UpdateSlot(appId, slotId string, slot *Slot) error {
	logrus.Debug("ListVersions from DummyStore")
	return nil
}
func (dummy *DummyStore) DeleteSlot(appId, slotId string) error {
	logrus.Debug("ListVersions from DummyStore")
	return nil
}

func (dummy *DummyStore) UpdateCurrentTask(appId, slotId string, task *Task) error {
	logrus.Debug("ListVersions from DummyStore")
	return nil
}
func (dummy *DummyStore) ListTaskHistory(appId, slotId string) []*Task {
	logrus.Debug("ListVersions from DummyStore")
	return nil
}

func (dummy *DummyStore) UpdateFrameworkId(frameworkId string) error {
	logrus.Debug("ListVersions from DummyStore")
	return nil
}
func (dummy *DummyStore) GetFrameworkId() string {
	logrus.Debug("ListVersions from DummyStore")
	return ""
}

func (dummy *DummyStore) CreateOfferAllocatorItem(item *OfferAllocatorItem) error {
	logrus.Debug("ListVersions from DummyStore")
	return nil
}

func (dummy *DummyStore) DeleteOfferAllocatorItem(slotId string) error {
	logrus.Debug("DeleteOfferAllocatorItem from DummyStore")
	return nil
}

func (dummy *DummyStore) ListOfferallocatorItems() []*OfferAllocatorItem {
	logrus.Debug("ListOfferallocatorItems from DummyStore")
	return nil
}

func (dummy *DummyStore) Synchronize() error {
	logrus.Debug("Synchronize from DummyStore")
	return nil
}
