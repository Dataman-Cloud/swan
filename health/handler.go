package health

import (
	"github.com/Dataman-Cloud/swan/types"
)

func (hc *HealthChecker) HealthCheckFailedHandler(check *types.Check, queue chan types.ReschedulerMsg) error {
	errC := make(chan error)
	msg := types.ReschedulerMsg{
		AppID:  check.AppID,
		TaskID: check.TaskID,
		Err:    errC,
	}

	queue <- msg

	return <-msg.Err
}
