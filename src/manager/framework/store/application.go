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

// To update an application version we need the follow steps in one transaction
// 1. find the old app info from database.
// 2. set new version's pervious versionId to the old version's id
// 3. push thie old version to version history
// 4. store the new version in app data
// 5. put all actions in one storeActions to propose data.
func (s *FrameworkStore) UpdateAppVersion(ctx context.Context, appId string, version *types.Version, cb func()) error {
	app, err := s.GetApp(appId)
	if err != nil {
		return err
	}

	if app == nil {
		return errors.New("Update app failed: target app was not found")
	}

	var storeActions []*types.StoreAction
	updateVersionAction := &types.StoreAction{
		Action: types.StoreActionKindCreate,
		Target: &types.StoreAction_Version{app.Version},
	}
	storeActions = append(storeActions, updateVersionAction)

	version.PerviousVersionID = app.Version.ID
	app.Version = version
	updateAppAction := &types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Application{app},
	}
	storeActions = append(storeActions, updateAppAction)

	return s.RaftNode.ProposeValue(ctx, storeActions, cb)
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
