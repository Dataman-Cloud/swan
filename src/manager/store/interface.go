package store

import (
	"golang.org/x/net/context"
)

type Store interface {
	CreateApp(app *Application) error
	UpdateApp(app *Application) error
	GetApp(appId string) *Application
	ListApps() []*Application
	DeleteApp(appId string) error

	CreateVersion(appId string, version *Version) error
	GetVersion(appId, versionId string) *Version
	ListVersions(appId string) []*Version

	CreateSlot(slot *Slot) error
	GetSlot(appId, slotId string) *Slot
	ListSlots(appId string) []*Slot
	UpdateSlot(appId, slotId string, slot *Slot) error
	DeleteSlot(appId, slotId string) error

	UpdateCurrentTask(appId, slotId string, task *Task) error
	ListTaskHistory(appId, slotId string) []*Task

	UpdateFrameworkId(frameworkId string) error
	GetFrameworkId() string

	CreateOfferAllocatorItem(item *OfferAllocatorItem) error
	DeleteOfferAllocatorItem(slotId string) error
	ListOfferallocatorItems() []*OfferAllocatorItem

	Recover() error
	Start(context.Context) error
}
