package store

import (
	raftstore "github.com/Dataman-Cloud/swan/src/manager/raft/store"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/boltdb/bolt"

	"golang.org/x/net/context"
)

func (store *ManagerStore) SaveApplication(application *types.Application) error {
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Application{application},
	}}

	return store.RaftNode.ProposeValue(context.TODO(), storeActions, nil)
}

func (store *ManagerStore) FetchApplication(appId string) (*types.Application, error) {
	app := &types.Application{}

	if err := store.BoltbDb.View(func(tx *bolt.Tx) error {
		return raftstore.WithAppBucket(tx, appId, func(bkt *bolt.Bucket) error {
			p := bkt.Get(raftstore.BucketKeyData)

			return app.Unmarshal(p)
		})
	}); err != nil {
		return nil, err
	}

	return app, nil
}

func (store *ManagerStore) ListApplications() ([]*types.Application, error) {
	var apps []*types.Application

	if err := store.BoltbDb.View(func(tx *bolt.Tx) error {
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

func (store *ManagerStore) DeleteApplication(appId string) error {
	removeApp := &types.Application{ID: appId}
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_Application{removeApp},
	}}

	return store.RaftNode.ProposeValue(context.TODO(), storeActions, nil)
}

func (store *ManagerStore) IncreaseApplicationUpdatedInstances(appId string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.UpdatedInstances += 1

	return store.SaveApplication(app)
}

func (store *ManagerStore) IncreaseApplicationInstances(appId string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.Instances += 1

	return store.SaveApplication(app)
}

func (store *ManagerStore) ResetApplicationUpdatedInstances(appId string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.UpdatedInstances = 0

	return store.SaveApplication(app)
}

func (store *ManagerStore) UpdateApplicationStatus(appId, status string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.Status = status

	return store.SaveApplication(app)
}

func (store *ManagerStore) IncreaseApplicationRunningInstances(appId string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.RunningInstances += 1

	return store.SaveApplication(app)
}

func (store *ManagerStore) ReduceApplicationRunningInstances(appId string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.RunningInstances -= 1

	return store.SaveApplication(app)
}

func (store *ManagerStore) ReduceApplicationInstances(appId string) error {
	app, err := store.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.Instances -= 1

	return store.SaveApplication(app)
}
