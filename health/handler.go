package health

import (
	"github.com/Dataman-Cloud/swan/types"
)

type HandlerFunc func(string, string) error

func (hc *HealthChecker) HealthCheckFailedHandler(appId, taskId string) error {
	errC := make(chan error)
	msg := types.ReschedulerMsg{
		AppID:  appId,
		TaskID: taskId,
		Err:    errC,
	}

	hc.msgQueue <- msg

	return <-msg.Err
}
