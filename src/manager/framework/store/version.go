package store

import (
	"errors"

	raftstore "github.com/Dataman-Cloud/swan/src/manager/raft/store"
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
	"golang.org/x/net/context"
)

// To update an application version we need the follow steps in one transaction
// 1. find the old app info from database.
// 2. set new version's pervious versionId to the old version's id
// 3. push thie old version to version history
// 4. store the new version in app data
// 5. put all actions in one storeActions to propose data.
func (s *FrameworkStore) UpdateVersion(ctx context.Context, appId string, version *types.Version, cb func()) error {
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

// retuns app history versions
func (s *FrameworkStore) ListVersions(appId string) ([]*types.Version, error) {
	var versions []*types.Version

	if err := s.BoltbDb.View(func(tx *bolt.Tx) error {
		bkt := raftstore.GetVersionsBucket(tx, appId)
		if bkt == nil {
			versions = []*types.Version{}
			return nil
		}

		return bkt.ForEach(func(k, v []byte) error {
			versionsBkt := raftstore.GetVersionBucket(tx, appId, string(k))
			if versionsBkt == nil {
				return nil
			}

			version := &types.Version{}
			p := versionsBkt.Get(raftstore.BucketKeyData)
			if err := version.Unmarshal(p); err != nil {
				return err
			}

			versions = append(versions, version)
			return nil
		})
	}); err != nil {
		return nil, err
	}

	return versions, nil
}
