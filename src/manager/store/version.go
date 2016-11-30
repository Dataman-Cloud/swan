package store

import (
	"fmt"
	"time"

	raftstore "github.com/Dataman-Cloud/swan/src/manager/raft/store"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/boltdb/bolt"

	"golang.org/x/net/context"
)

func (store *ManagerStore) SaveVersion(version *types.Version) error {
	version.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindUpdate,
		Target: &types.StoreAction_Version{version},
	}}

	return store.RaftNode.ProposeValue(context.TODO(), storeActions, nil)
}

func (store *ManagerStore) ListVersionId(appId string) ([]string, error) {
	var versionIdList []string

	if err := store.BoltbDb.View(func(tx *bolt.Tx) error {
		bkt := raftstore.GetVersionsBucket(tx, appId)
		if bkt == nil {
			versionIdList = []string{}
			return nil
		}

		return bkt.ForEach(func(k, v []byte) error {
			versionBkt := raftstore.GetVersionBucket(tx, appId, string(k))
			if versionBkt == nil {
				return nil
			}

			versionIdList = append(versionIdList, string(k))
			return nil
		})

	}); err != nil {
		return nil, err
	}

	return versionIdList, nil
}

func (store *ManagerStore) FetchVersion(appId, versionId string) (*types.Version, error) {
	version := &types.Version{}

	if err := store.BoltbDb.View(func(tx *bolt.Tx) error {
		return raftstore.WithVersionBucket(tx, appId, versionId, func(bkt *bolt.Bucket) error {
			p := bkt.Get(raftstore.BucketKeyData)

			return version.Unmarshal(p)
		})

	}); err != nil {
		return nil, err
	}

	return version, nil
}

func (store *ManagerStore) DeleteVersion(appId, versionId string) error {
	version := &types.Version{ID: versionId, AppId: appId}
	storeActions := []*types.StoreAction{&types.StoreAction{
		Action: types.StoreActionKindRemove,
		Target: &types.StoreAction_Version{version},
	}}

	return store.RaftNode.ProposeValue(context.TODO(), storeActions, nil)
}
