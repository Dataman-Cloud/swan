package health

import (
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
)

type HandlerFunc func(string, string) error

func (m *HealthCheckManager) HealthCheckFailedHandler(appId, taskId string) error {
	task, err := m.store.FetchApplicationTask(appId, taskId)
	if err != nil {
		return err
	}

	app, err := m.store.FetchApplication(appId)
	if err != nil {
		return err
	}

	if task.Status == "RESCHEDULING" || app.Status == "UPDATING" {
		return nil
	} else {
		if err := m.store.UpdateTask(appId, taskId, "RESCHEDULING"); err != nil {
			logrus.Errorf("Updating task status failed for health check failover: %s", err.Error())
		}
	}

	msg := types.ReschedulerMsg{
		AppID:  appId,
		TaskID: taskId,
		Err:    make(chan error),
	}

	m.msgQueue <- msg

	return <-msg.Err
}
