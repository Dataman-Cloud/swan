package boltdb

import (
	"github.com/Dataman-Cloud/swan/types"
	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
)

func (db *Boltdb) PutHealthcheck(task *types.Task, port int64, appId string) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, healthCheck := range task.HealthChecks {

		check := types.Check{
			ID:          task.Name,
			Address:     task.AgentHostname,
			Port:        port,
			TaskID:      task.Name,
			AppID:       appId,
			Protocol:    healthCheck.Protocol,
			Interval:    healthCheck.IntervalSeconds,
			Timeout:     healthCheck.TimeoutSeconds,
			Command:     healthCheck.Command,
			Path:        healthCheck.Path,
			MaxFailures: healthCheck.MaxConsecutiveFailures,
		}

		if err := withCreateHealthcheckBucketIfNotExists(tx, check.AppID, check.ID, func(bkt *bolt.Bucket) error {
			p, err := proto.Marshal(&check)
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

func (db *Boltdb) GetHealthChecks(appId string) ([]*types.Check, error) {
	var checks []*types.Check

	if err := db.View(func(tx *bolt.Tx) error {
		bkt := getHealthChecksBucket(tx, appId)
		if bkt == nil {
			checks = []*types.Check{}
			return nil
		}

		if err := bkt.ForEach(func(k, v []byte) error {
			healthCheckBkt := bkt.Bucket(k)
			if healthCheckBkt == nil {
				return nil
			}

			p := healthCheckBkt.Get(bucketKeyData)
			var check types.Check
			if err := proto.Unmarshal(p, &check); err != nil {
				return err
			}

			checks = append(checks, &check)
			return nil

		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return checks, nil
}

func (db *Boltdb) DeleteHealthCheck(appId, healthCheckId string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bkt := getHealthChecksBucket(tx, appId)
		if bkt == nil {
			return nil
		}

		return bkt.DeleteBucket([]byte(healthCheckId))
	})
}
