package store

import (
	"errors"

	raftstore "github.com/Dataman-Cloud/swan/src/manager/raft/store"
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"
	"github.com/boltdb/bolt"

	"golang.org/x/net/context"
)

func (s *FrameworkStore) CreateApp(ctx context.Context, app *types.Application, cb func()) error {
	storeAction := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Application{app},
	}}

	return s.RaftNode.ProposeValue(ctx, storeAction, cb)
}

func (s *FrameworkStore) UpdateApp(ctx context.Context, app *types.Application, cb func()) error {
	storeAction := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Application{app},
	}}

	return s.RaftNode.ProposeValue(ctx, storeAction, cb)
}

func (s *FrameworkStore) UpdateAppState(ctx context.Context, appId, state string, cb func()) error {
	app, err := s.GetApp(appId)
	if err != nil {
		return err
	}

	if app == nil {
		return ErrAppNotFound
	}

	app.State = state

	return s.UpdateApp(ctx, app, cb)
}

func (s *FrameworkStore) GetApp(appId string) (*types.Application, error) {
	app := &types.Application{}

	if err := s.BoltbDb.View(func(tx *bolt.Tx) error {
		return raftstore.WithAppBucket(tx, appId, func(bkt *bolt.Bucket) error {
			p := bkt.Get(raftstore.BucketKeyData)

			return app.Unmarshal(p)
		})
	}); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *FrameworkStore) ListApps() ([]*types.Application, error) {
	var apps []*types.Application

	if err := s.BoltbDb.View(func(tx *bolt.Tx) error {
		bkt := raftstore.GetAppsBucket(tx)
		if bkt == nil {
			apps = []*types.Application{}
			return nil
		}

		return bkt.ForEach(func(k, v []byte) error {
			appBucket := raftstore.GetAppBucket(tx, string(k))
			if appBucket == nil {
				return nil
			}

			app := &types.Application{}
			p := appBucket.Get(raftstore.BucketKeyData)
			if err := app.Unmarshal(p); err != nil {
				return err
			}

			apps = append(apps, app)
			return nil

		})

	}); err != nil {
		return nil, err
	}

	return apps, nil
}

func (s *FrameworkStore) DeleteApp(ctx context.Context, appId string, cb func()) error {
	removeApp := &types.Application{ID: appId}
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_Application{removeApp},
	}}

	return s.RaftNode.ProposeValue(ctx, storeActions, cb)
}

func (s *FrameworkStore) CommitAppProposeVersion(ctx context.Context, app *types.Application, cb func()) error {
	if app.ProposedVersion == nil {
		return errors.New("commit propose version failed: propose version was nil")
	}

	var storeActions []*types.StoreAction
	updateVersionAction := &types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Version{app.Version},
	}
	storeActions = append(storeActions, updateVersionAction)

	app.Version = app.ProposedVersion
	app.ProposedVersion = nil

	updateAppAction := &types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Application{app},
	}
	storeActions = append(storeActions, updateAppAction)

	return s.RaftNode.ProposeValue(ctx, storeActions, cb)
}
