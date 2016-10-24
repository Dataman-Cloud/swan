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

	if task.Status == "RESCHEDULING" {
		return nil
	} else {
		task.Status = "RESCHEDULING"
		if err := m.store.RegisterTask(task); err != nil {
			logrus.Errorf("Register task failed: %s", err.Error())
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
