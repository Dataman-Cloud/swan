package store

import (
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

func (s *FrameworkStore) ListApplications() ([]*types.Application, error) {
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

func (s *FrameworkStore) DeleteApplication(ctx context.Context, appId string, cb func()) error {
	removeApp := &types.Application{ID: appId}
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_Application{removeApp},
	}}

	return s.RaftNode.ProposeValue(ctx, storeActions, cb)
}
