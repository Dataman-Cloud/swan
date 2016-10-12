package consul

import (
	"errors"
	"fmt"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	consul "github.com/hashicorp/consul/api"
	"gopkg.in/mgo.v2/bson"
)

type Consul struct {
	client *consul.Client
}

func NewConsul(addr string) (*Consul, error) {
	cfg := consul.Config{
		Address: addr,
		Scheme:  "http",
	}

	client, err := consul.NewClient(&cfg)
	if err != nil {
		return nil, err
	}

	return &Consul{
		client: client,
	}, nil
}

// RegisterFrameworkID is used to register the frameworkId in consul.
// This ID is used for framework re-register.
func (c *Consul) RegisterFrameworkID(frameworkId string) error {
	logrus.WithFields(logrus.Fields{"FrameworkID": frameworkId}).Info("Register frameworkId in consul")

	kv := consul.KVPair{
		Key:   "swan/frameworkid",
		Value: []byte(frameworkId),
	}

	_, err := c.client.KV().Put(&kv, nil)
	if err != nil {
		return err
	}

	return nil
}

// FetchFrameworkID is used to fetch frameworkId from consul if exists for framework registration with mesos.
func (c *Consul) FetchFrameworkID(namespace string) (string, error) {
	frameworkId, _, err := c.client.KV().Get("swan/frameworkid", nil)
	if err != nil {
		return "", err
	}

	if frameworkId == nil {
		return "", nil
	}

	return string(frameworkId.Value[:]), nil
}

// FrameworkIDHasRegistered is used to check whether the frameworkId has registered in consul.
// If has registered, return true; otherwise return false.
func (c *Consul) FrameworkIDHasRegistered(frameworkId string) (bool, error) {
	kv, _, err := c.client.KV().Get("swan/frameworkid", nil)
	if err != nil {
		logrus.Errorf("Fetch framework id from consul failed: %s", err)
		return false, err
	}

	if kv != nil {
		return true, nil
	}

	return false, nil
}

// RegisterApplication is used to register a application in consul. Use cluster_id/user_id
// as key, and application information as value.
func (c *Consul) RegisterApplication(application *types.Application) error {
	logrus.Infof("Register application %s in consul", application.ID)

	data, err := bson.Marshal(application)
	if err != nil {
		logrus.Infof("Marshal application failed: %s", err.Error())
		return err
	}

	app := consul.KVPair{
		//Key:   fmt.Sprintf("%s/%s/%s", application.ClusterID, application.UserID, application.ID),
		Key:   fmt.Sprintf("applications/%s", application.ID),
		Value: data,
	}

	_, err = c.client.KV().Put(&app, nil)
	if err != nil {
		logrus.Info("Register application %s in consul failed: %s", application.ID, err.Error())
		return err
	}

	return nil
}

// FetchApplication is used to fetch application from consul by application id. If no application
// found, nil will returned.
func (c *Consul) FetchApplication(id string) (*types.Application, error) {
	logrus.Infof("Fetch applicaiton %s from consul", id)

	app, _, err := c.client.KV().Get(fmt.Sprintf("applications/%s", id), nil)
	if err != nil {
		logrus.Errorf("Fetch appliction failed: %s", err.Error())
		return nil, err
	}

	if app == nil {
		logrus.Errorf("Application %s not found in consul", id)
		return nil, nil
	}

	var application types.Application
	if err := bson.Unmarshal(app.Value, &application); err != nil {
		logrus.Errorf("Unmarshal application failed: %s", err.Error())
		return nil, err
	}

	return &application, nil
}

// ListApplications is used to list all applications belong to a cluster or a user.
func (c *Consul) ListApplications() ([]*types.Application, error) {
	apps, _, err := c.client.KV().List("applications", nil)
	if err != nil {
		logrus.Errorf("Fetch appliction failed: %s", err.Error())
		return nil, err
	}

	var applications []*types.Application
	for _, app := range apps {
		var application types.Application

		if err := bson.Unmarshal(app.Value, &application); err != nil {
			logrus.Errorf("Unmarshal application failed: %s", err.Error())
			return nil, err
		}
		applications = append(applications, &application)
	}

	return applications, nil
}

// DeleteApplication is used to delete application from consul by application id.
func (c *Consul) DeleteApplication(id string) error {
	_, err := c.client.KV().Delete(fmt.Sprintf("applications/%s", id), nil)
	if err != nil {
		logrus.Errorf("Delete application %s failed: %s", id, err.Error())
		return err
	}

	return nil
}

// RegisterTask is used to register task in consul under task.AppId namespace.
func (c *Consul) RegisterTask(task *types.Task) error {
	data, err := bson.Marshal(task)
	if err != nil {
		logrus.Infof("Marshal task failed: %s", err.Error())
		return err
	}

	t := consul.KVPair{
		Key:   fmt.Sprintf("applications/%s/%s", task.AppId, task.ID),
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

		if err := bson.Unmarshal(task.Value, &t); err != nil {
			logrus.Errorf("Unmarshal application failed: %s", err.Error())
			return nil, err
		}
		tasksList = append(tasksList, &t)
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
	t, _, err := c.client.KV().Get(fmt.Sprintf("applications/%s/%s", applicationId, taskId), nil)
	if err != nil {
		logrus.Errorf("Fetch appliction failed: %s", err.Error())
		return nil, err
	}

	if t == nil {
		logrus.Errorf("Task %s not found in consul", taskId)
		return nil, errors.New("Task not found")
	}

	var task types.Task
	if err := bson.Unmarshal(t.Value, &task); err != nil {
		logrus.Errorf("Unmarshal application failed: %s", err.Error())
		return nil, err
	}

	return &task, nil
}
