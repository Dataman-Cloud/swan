package boltdb

import (
	"encoding/json"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
)

func (b *BoltStore) SaveApplication(application *types.Application) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("applications"))

	data, err := json.Marshal(application)
	if err != nil {
		logrus.Errorf("Marshal application failed: %s", err.Error())
		return err
	}

	if err := bucket.Put([]byte(application.ID), data); err != nil {
		return err
	}

	return tx.Commit()
}

func (b *BoltStore) FetchApplication(appId string) (*types.Application, error) {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("applications"))
	data := bucket.Get([]byte(appId))
	if data == nil {
		return nil, nil
	}

	var application types.Application
	if err := json.Unmarshal(data, &application); err != nil {
		logrus.Errorf("Unmarshal application failed: %s", err.Error())
		return nil, err
	}

	return &application, nil
}

func (b *BoltStore) ListApplications() ([]*types.Application, error) {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("applications"))

	appList := make([]*types.Application, 0)

	if err := bucket.ForEach(func(k, v []byte) error {
		var app types.Application
		if err := json.Unmarshal(v, &app); err != nil {
			return err
		}
		appList = append(appList, &app)

		return nil
	}); err != nil {
		return nil, err
	}

	return appList, nil
}

func (b *BoltStore) DeleteApplication(appId string) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("applications"))

	if err := bucket.Delete([]byte(appId)); err != nil {
		return err
	}

	return tx.Commit()
}

func (b *BoltStore) IncreaseApplicationUpdatedInstances(appId string) error {
	app, err := b.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.UpdatedInstances += 1

	err = b.SaveApplication(app)
	if err != nil {
		return err
	}

	return nil
}

func (b *BoltStore) IncreaseApplicationInstances(appId string) error {
	app, err := b.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.Instances += 1

	err = b.SaveApplication(app)
	if err != nil {
		return err
	}

	return nil
}

func (b *BoltStore) ResetApplicationUpdatedInstances(appId string) error {
	app, err := b.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.UpdatedInstances = 0

	err = b.SaveApplication(app)
	if err != nil {
		return err
	}

	return nil
}

func (b *BoltStore) UpdateApplicationStatus(appId, status string) error {
	app, err := b.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.Status = status

	err = b.SaveApplication(app)
	if err != nil {
		return err
	}
	return nil
}

func (b *BoltStore) IncreaseApplicationRunningInstances(appId string) error {
	app, err := b.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.RunningInstances += 1

	err = b.SaveApplication(app)
	if err != nil {
		return err
	}

	return nil
}

func (b *BoltStore) ReduceApplicationRunningInstances(appId string) error {
	app, err := b.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.RunningInstances -= 1

	err = b.SaveApplication(app)
	if err != nil {
		return err
	}

	return nil
}

func (b *BoltStore) ReduceApplicationInstances(appId string) error {
	app, err := b.FetchApplication(appId)
	if err != nil {
		return err
	}

	app.Instances -= 1

	err = b.SaveApplication(app)
	if err != nil {
		return err
	}

	return nil
}
