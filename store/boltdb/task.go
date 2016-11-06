package boltdb

import (
	"encoding/json"
	"errors"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
)

func (b *BoltStore) SaveTask(task *types.Task) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("tasks"))

	data, err := json.Marshal(task)
	if err != nil {
		logrus.Errorf("Marshal application failed: %s", err.Error())
		return err
	}

	if err := bucket.Put([]byte(task.Name), data); err != nil {
		return err
	}

	return tx.Commit()

}

func (b *BoltStore) ListTasks(applicationId string) ([]*types.Task, error) {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("tasks"))

	var tasksList []*types.Task

	bucket.ForEach(func(k, v []byte) error {
		var task types.Task
		if err := json.Unmarshal(v, &task); err != nil {
			return err
		}
		if task.AppId == applicationId {
			tasksList = append(tasksList, &task)
		}

		return nil
	})

	return tasksList, nil
}

func (b *BoltStore) DeleteApplicationTasks(appId string) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("tasks"))

	bucket.ForEach(func(k, v []byte) error {
		var task types.Task
		if err := json.Unmarshal(v, &task); err != nil {
			return err
		}
		if task.AppId == appId {
			if err := bucket.Delete([]byte(task.Name)); err != nil {
				return err
			}
		}

		return nil
	})

	return tx.Commit()
}

func (b *BoltStore) DeleteTask(taskId string) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("tasks"))

	if err := bucket.Delete([]byte(taskId)); err != nil {
		return err
	}

	return tx.Commit()
}

func (b *BoltStore) FetchTask(taskId string) (*types.Task, error) {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("tasks"))

	data := bucket.Get([]byte(taskId))
	if data == nil {
		return nil, errors.New("Not Found")
	}

	var task types.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

func (b *BoltStore) UpdateTaskStatus(taskId string, status string) error {
	task, err := b.FetchTask(taskId)
	if err != nil {
		return err
	}

	task.Status = status

	if err := b.SaveTask(task); err != nil {
		return err
	}

	return nil
}
