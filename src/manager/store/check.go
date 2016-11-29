package store

import (
	"encoding/json"

	"github.com/Dataman-Cloud/swan/src/types"
)

func (store *ManagerStore) SaveCheck(task *types.Task, port uint32, appId string) error {
	tx, err := store.BoltbDb.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("checks"))

	for _, healthCheck := range task.HealthChecks {

		check := types.Check{
			ID:       task.Name,
			Address:  *task.AgentHostname,
			Port:     int(port),
			TaskID:   task.Name,
			AppID:    appId,
			Protocol: healthCheck.Protocol,
			Interval: int(healthCheck.IntervalSeconds),
			Timeout:  int(healthCheck.TimeoutSeconds),
		}

		// TODO(pwzgorilla clear unuse code)
		//if healthCheck.Command != nil {
		//	check.Command = healthCheck.Command
		//}

		//if healthCheck.Path != nil {
		//	check.Path = *healthCheck.Path
		//}

		//if healthCheck.ConsecutiveFailures != 0 {
		//	check.MaxFailures = int(healthCheck.ConsecutiveFailures)
		//}

		data, err := json.Marshal(&check)
		if err != nil {
			return err
		}

		if err := bucket.Put([]byte(check.ID), data); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (store *ManagerStore) ListChecks() ([]*types.Check, error) {
	tx, err := store.BoltbDb.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("checks"))

	var checks []*types.Check
	if err := bucket.ForEach(func(k, v []byte) error {
		var check types.Check
		if err := json.Unmarshal(v, &check); err != nil {
			return err
		}
		checks = append(checks, &check)

		return nil
	}); err != nil {
		return nil, err
	}

	return checks, nil
}

func (store *ManagerStore) DeleteCheck(checkId string) error {
	tx, err := store.BoltbDb.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("checks"))

	if err := bucket.Delete([]byte(checkId)); err != nil {
		return err
	}

	return tx.Commit()
}
