package consul

import (
	"github.com/Sirupsen/logrus"
	consul "github.com/hashicorp/consul/api"
)

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
