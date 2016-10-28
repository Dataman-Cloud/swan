package health

import (
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
)

type HandlerFunc func(string, string) error

func (m *HealthCheckManager) HealthCheckFailedHandler(appId, taskId string) error {
	logrus.Infof("Reschduler task %s for health check failed", taskId)
	msg := types.ReschedulerMsg{
		AppID:  appId,
		TaskID: taskId,
		Err:    make(chan error),
	}

	m.msgQueue <- msg

	return <-msg.Err
}
