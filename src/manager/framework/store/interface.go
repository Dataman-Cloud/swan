package store

import (
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"golang.org/x/net/context"
)

type Store interface {
	CreateApp(ctx context.Context, app *types.Application, cb func()) error
	UpdateApp(ctx context.Context, app *types.Application, cb func()) error
	UpdateAppState(ctx context.Context, appId, state string, cb func()) error
	CommitAppProposeVersion(ctx context.Context, app *types.Application, cb func()) error
	GetApp(appId string) (*types.Application, error)
	ListApps() ([]*types.Application, error)
	DeleteApp(ctx context.Context, appId string, cb func()) error
	UpdateVersion(ctx context.Context, appId string, version *types.Version, cb func()) error
	GetVersion(appId, versionId string) (*types.Version, error)
	ListVersions(appId string) ([]*types.Version, error)
	CreateSlot(ctx context.Context, slot *types.Slot, cb func()) error
	GetSlot(appId, slotId string) (*types.Slot, error)
	ListSlots(appId string) ([]*types.Slot, error)
	UpdateSlot(ctx context.Context, slot *types.Slot, cb func()) error
	DeleteSlot(ctx context.Context, appId, slotId string, cb func()) error
	UpdateTask(ctx context.Context, task *types.Task, cb func()) error
	ListTasks(appId, slotId string) ([]*types.Task, error)
	UpdateFrameworkId(ctx context.Context, frameworkId string, cb func()) error
	GetFrameworkId() (string, error)

	CreateOfferAllocatorItem(context.Context, *types.OfferAllocatorItem, func()) error
	DeleteOfferAllocatorItem(context.Context, string, func()) error
	ListOfferallocatorItems() ([]*types.OfferAllocatorItem, error)
}
