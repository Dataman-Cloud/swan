package store

import (
	"errors"

	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"golang.org/x/net/context"
)

type slotStore struct {
	Body          []byte
	CurrentTask   [][]byte
	TaskHistories [][]byte
}

type appStore struct {
	Versions map[string][]byte
	Slots    map[string]slotStore
}

type ZkStore struct {
	m map[string]interface{}
}

func NewZkStore() *ZkStore {
	return &ZkStore{
		m: make(map[string]interface{}),
	}
}

func (zk *ZkStore) CreateApp(ctx context.Context, app *types.Application, cb func()) error {
	if _, found := zk.m[app.ID]; found {
		return errors.New("app exists")
	}

	zk.m[app.ID] = app

	return nil
}

func (zk *ZkStore) UpdateApp(ctx context.Context, app *types.Application, cb func()) error {
	return nil
}

func (zk *ZkStore) GetApp(appId string) (*types.Application, error) {
	return nil, nil
}

func (zk *ZkStore) ListApps() ([]*types.Application, error) {
	return nil, nil
}

func (zk *ZkStore) DeleteApp(ctx context.Context, appId string, cb func()) error {
	return nil
}

func (zk *ZkStore) CreateVersion(ctx context.Context, appId string, version *types.Version, cb func()) error {
	return nil
}

func (zk *ZkStore) GetVersion(appId, versionId string) (*types.Version, error) {
	return nil, nil
}

func (zk *ZkStore) ListVersions(appId string) ([]*types.Version, error) {
	return nil, nil
}

func (zk *ZkStore) CreateSlot(ctx context.Context, slot *types.Slot, cb func()) error {
	return nil
}

func (zk *ZkStore) GetSlot(appId, slotId string) (*types.Slot, error) {
	return nil, nil
}

func (zk *ZkStore) ListSlots(appId string) ([]*types.Slot, error) {
	return nil, nil
}

func (zk *ZkStore) UpdateSlot(ctx context.Context, slot *types.Slot, cb func()) error {
	return nil
}

func (zk *ZkStore) DeleteSlot(ctx context.Context, appId, slotId string, cb func()) error {
	return nil
}

func (zk *ZkStore) UpdateTask(ctx context.Context, task *types.Task, cb func()) error {
	return nil
}

func (zk *ZkStore) ListTasks(appId, slotId string) ([]*types.Task, error) {
	return nil, nil
}

func (zk *ZkStore) UpdateFrameworkId(ctx context.Context, frameworkId string, cb func()) error {
	return nil
}

func (zk *ZkStore) GetFrameworkId() (string, error) {
	return "", nil
}
func (zk *ZkStore) CreateOfferAllocatorItem(context.Context, *types.OfferAllocatorItem, func()) error {
	return nil
}
func (zk *ZkStore) DeleteOfferAllocatorItem(context.Context, string, func()) error {
	return nil
}
func (zk *ZkStore) ListOfferallocatorItems() ([]*types.OfferAllocatorItem, error) {
	return nil, nil
}
