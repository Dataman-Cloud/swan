package store

import (
	"errors"

	"github.com/Dataman-Cloud/swan/src/types"
)

func (store *ManagerStore) SaveTask(task *types.Task) error {
	return nil
}

func (store *ManagerStore) ListTasks(appId string) ([]*types.Task, error) {
	return nil, nil
}

func (store *ManagerStore) DeleteTask(appId, taskId string) error {
	return nil
}

func (store *ManagerStore) FetchTask(appId, taskId string) (*types.Task, error) {
	return nil, nil
}

func (store *ManagerStore) UpdateTaskStatus(appId, taskId string, status string) error {
	task, err := store.FetchTask(appId, taskId)
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
