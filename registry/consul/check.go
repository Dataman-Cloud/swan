package consul

import (
	"encoding/json"
	"fmt"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	consul "github.com/hashicorp/consul/api"
)

// RegisterCheck register a check in consul for health check.
func (c *Consul) RegisterCheck(task *types.Task, port uint32, appId string) error {
	logrus.Info("Register check in consul")

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

		if healthCheck.Command != nil {
			check.Command = healthCheck.Command
		}

		if healthCheck.Path != nil {
			check.Path = *healthCheck.Path
		}

		if healthCheck.MaxConsecutiveFailures != nil {
			check.MaxFailures = *healthCheck.MaxConsecutiveFailures
		}

		data, err := json.Marshal(&check)
		if err != nil {
			logrus.Info("Marshal check failed: %s", err.Error())
			return err
		}

		s := consul.KVPair{
			Key:   fmt.Sprintf("checks/%s", task.Name),
			Value: data,
		}

		if _, err := c.client.KV().Put(&s, nil); err != nil {
			logrus.Errorf("Register check %s in consul failed: %s", task.Name, err.Error())
			return err
		}
	}

	return nil
}

// ListChecks list all checks in consul.
func (c *Consul) ListChecks() ([]*types.Check, error) {
	cs, _, err := c.client.KV().List("checks", nil)
	if err != nil {
		logrus.Errorf("List checks failed: %s")
	}

	var checks []*types.Check
	for _, c := range cs {
		var check types.Check

		if err := json.Unmarshal(c.Value, &check); err != nil {
			logrus.Errorf("Unmarshal check failed: %s", err.Error())
			return nil, err
		}

		checks = append(checks, &check)
	}

	return checks, nil
}

// DeleteCheck delete specified check by check id.
func (c *Consul) DeleteCheck(checkId string) error {
	_, err := c.client.KV().Delete(fmt.Sprintf("checks/%s", checkId), nil)
	if err != nil {
		logrus.Errorf("Delete check %s failed: %s", checkId, err.Error())
		return err
	}

	return nil
}
