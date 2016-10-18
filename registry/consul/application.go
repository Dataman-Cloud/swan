package consul

import (
	"encoding/json"
	"fmt"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	consul "github.com/hashicorp/consul/api"
)

// RegisterApplication is used to register a application in consul. Use cluster_id/user_id
// as key, and application information as value.
func (c *Consul) RegisterApplication(application *types.Application) error {
	logrus.Infof("Register application %s in consul", application.ID)

	data, err := json.Marshal(application)
	if err != nil {
		logrus.Infof("Marshal application failed: %s", err.Error())
		return err
	}

	app := consul.KVPair{
		Key:   fmt.Sprintf("applications/%s/info", application.ID),
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

	app, _, err := c.client.KV().Get(fmt.Sprintf("applications/%s/info", id), nil)
	if err != nil {
		logrus.Errorf("Fetch appliction failed: %s", err.Error())
		return nil, err
	}

	if app == nil {
		logrus.Errorf("Application %s not found in consul", id)
		return nil, nil
	}

	var application types.Application
	if err := json.Unmarshal(app.Value, &application); err != nil {
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

		if err := json.Unmarshal(app.Value, &application); err != nil {
			logrus.Errorf("Unmarshal application failed: %s", err.Error())
			return nil, err
		}
		applications = append(applications, &application)
	}

	return applications, nil
}

// DeleteApplication is used to delete application from consul by application id.
func (c *Consul) DeleteApplication(id string) error {
	_, err := c.client.KV().DeleteTree(fmt.Sprintf("applications/%s", id), nil)
	if err != nil {
		logrus.Errorf("Delete application %s failed: %s", id, err.Error())
		return err
	}

	return nil
}

// UpdateApplication is used to update application's instance count.
func (c *Consul) UpdateApplication(application *types.Application) error {
	return c.RegisterApplication(application)
}
