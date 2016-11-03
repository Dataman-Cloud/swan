package boltdb

import (
	"github.com/Dataman-Cloud/swan/types"
	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
)

func (db *Boltdb) PutApp(app *types.Application) error {
	return db.PutApps(app)
}

func (db *Boltdb) PutApps(apps ...*types.Application) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, app := range apps {
		if err := withCreateAppBucketIfNotExists(tx, app.ID, func(bkt *bolt.Bucket) error {
			p, err := proto.Marshal(app)
			if err != nil {
				return err
			}

			return bkt.Put(bucketKeyData, p)
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *Boltdb) GetApp(appId string) (*types.Application, error) {
	apps, err := db.GetApps(appId)
	if err != nil {
		return nil, err
	}

	if len(apps) < 1 {
		return nil, errAppUnknown
	}

	return apps[0], nil
}

func (db *Boltdb) GetApps(appIds ...string) ([]*types.Application, error) {
	if appIds == nil {
		return db.getAllApps()
	}

	var apps []*types.Application
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}

	for _, appId := range appIds {
		if err := withAppBucket(tx, appId, func(bkt *bolt.Bucket) error {
			var app types.Application
			p := bkt.Get(bucketKeyData)
			if err := proto.Unmarshal(p, &app); err != nil {
				return err
			}

			apps = append(apps, &app)
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return apps, nil
}

func (db *Boltdb) getAllApps() ([]*types.Application, error) {
	var apps []*types.Application
	if err := db.View(func(tx *bolt.Tx) error {
		bkt := getAppsBucket(tx)
		if bkt == nil {
			apps = []*types.Application{}
			return nil
		}

		if err := bkt.ForEach(func(k, v []byte) error {
			appBucket := getAppBucket(tx, string(k))
			if appBucket == nil {
				return nil
			}

			var app types.Application
			p := appBucket.Get(bucketKeyData)
			if err := proto.Unmarshal(p, &app); err != nil {
				return err
			}

			apps = append(apps, &app)
			return nil

		}); err != nil {
			return err
		}

		return nil

	}); err != nil {
		return nil, err
	}

	return apps, nil
}

func (db *Boltdb) DeleteApp(appId string) error {
	return db.DeleteApps(appId)
}

func (db *Boltdb) DeleteApps(appIds ...string) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bkt := getAppsBucket(tx)
	if bkt == nil {
		return nil
	}

	for _, appId := range appIds {
		if err := bkt.Delete([]byte(appId)); err != nil {
			return err
		}
	}

	return tx.Commit()
}
