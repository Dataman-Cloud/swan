package consul

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	consul "github.com/hashicorp/consul/api"
)

// RegisterTask is used to register task in consul under task.AppId namespace.
func (c *Consul) RegisterTask(task *types.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		logrus.Infof("Marshal task failed: %s", err.Error())
		return err
	}

	t := consul.KVPair{
		Key:   fmt.Sprintf("applications/%s/tasks/%s", task.AppId, task.Name),
		Value: data,
	}

	_, _, err = c.client.KV().CAS(&t, nil)
	if err != nil {
		logrus.Info("Register task %s in consul failed: %s", task.ID, err.Error())
		return err
	}

	return nil
}

// ListApplicationTasks is used to get all tasks belong to a application from consul.
func (c *Consul) ListApplicationTasks(applicationId string) ([]*types.Task, error) {
	tasks, _, err := c.client.KV().List(fmt.Sprintf("applications/%s", applicationId), nil)
	if err != nil {
		logrus.Errorf("Fetch appliction tasks failed: %s", err.Error())
		return nil, err
	}

	var tasksList []*types.Task
	for _, task := range tasks {
		var t types.Task

		if err := json.Unmarshal(task.Value, &t); err != nil {
			logrus.Errorf("Unmarshal application failed: %s", err.Error())
			return nil, err
		}
		// Filter application
		if t.ID != applicationId {
			tasksList = append(tasksList, &t)
		}
	}

	return tasksList, nil
}

// DeleteApplicationTasks is used to delete all tasks belong to a application from consul.
func (c *Consul) DeleteApplicationTasks(applicationId string) error {
	_, err := c.client.KV().DeleteTree(fmt.Sprintf("applications/%s", applicationId), nil)
	if err != nil {
		logrus.Errorf("Delete tasks failed: %s", err.Error())
		return err
	}

	return nil
}

// DeleteApplicationTask is used to delete specified task belong to a application from consul.
func (c *Consul) DeleteApplicationTask(applicationId, taskId string) error {
	_, err := c.client.KV().Delete(fmt.Sprintf("applications/%s/tasks/%s", applicationId, taskId), nil)
	if err != nil {
		logrus.Errorf("Delete task %s failed: %s", taskId, err.Error())
		return err
	}

	return nil
}

// FetchApplicationTask is used to fetch a task belong to a application from consul.
func (c *Consul) FetchApplicationTask(applicationId, taskId string) (*types.Task, error) {
	t, _, err := c.client.KV().Get(fmt.Sprintf("applications/%s/tasks/%s", applicationId, taskId), nil)
	if err != nil {
		logrus.Errorf("Fetch appliction failed: %s", err.Error())
		return nil, err
	}

	if t == nil {
		logrus.Errorf("Task %s not found in consul", taskId)
		return nil, errors.New("Task not found")
	}

	var task types.Task
	if err := json.Unmarshal(t.Value, &task); err != nil {
		logrus.Errorf("Unmarshal application failed: %s", err.Error())
		return nil, err
	}

	return &task, nil
}
