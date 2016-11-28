package store

import (
	"encoding/json"
	"errors"

	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
)

func (store *ManagerStore) SaveTask(task *types.Task) error {
	tx, err := store.BoltbDb.Begin(true)
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

func (store *ManagerStore) ListTasks(applicationId string) ([]*types.Task, error) {
	tx, err := store.BoltbDb.Begin(true)
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

func (store *ManagerStore) DeleteApplicationTasks(appId string) error {
	tx, err := store.BoltbDb.Begin(true)
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

func (store *ManagerStore) DeleteTask(taskId string) error {
	tx, err := store.BoltbDb.Begin(true)
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

func (store *ManagerStore) FetchTask(taskId string) (*types.Task, error) {
	tx, err := store.BoltbDb.Begin(true)
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

func (store *ManagerStore) UpdateTaskStatus(taskId string, status string) error {
	task, err := store.FetchTask(taskId)
	if err != nil {
		return err
	}

	if task == nil {
		return errors.New("task not found")
	}

	task.Status = status

	if err := store.SaveTask(task); err != nil {
		return err
	}

	return nil
}
