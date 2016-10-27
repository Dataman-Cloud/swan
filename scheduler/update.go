package scheduler

import (
	"net/http"
	"strings"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
)

func (s *Scheduler) status(status *mesos.TaskStatus) {
	// send ack
	if status.GetUuid() != nil {
		call := &sched.Call{
			FrameworkId: s.framework.GetId(),
			Type:        sched.Call_ACKNOWLEDGE.Enum(),
			Acknowledge: &sched.Call_Acknowledge{
				AgentId: status.GetAgentId(),
				TaskId:  status.GetTaskId(),
				Uuid:    status.GetUuid(),
			},
		}

		// send call
		resp, err := s.send(call)
		if err != nil {
			logrus.Error("Unable to send Acknowledge Call: ", err)
			return
		}
		if resp.StatusCode != http.StatusAccepted {
			logrus.Error("Acknowledge call returned unexpected status: %d", resp.StatusCode)
		}
	}

	ID := status.TaskId.GetValue()
	state := status.GetState()

	taskId := strings.Split(ID, "-")[1]
	appId := strings.Split(taskId, ".")[1]

	var STATUS string

	switch state {
	case mesos.TaskState_TASK_STAGING:
		STATUS = "STAGING"
	case mesos.TaskState_TASK_STARTING:
		STATUS = "STARTING"
	case mesos.TaskState_TASK_RUNNING:
		logrus.Infof("Task %s is RUNNING", taskId)
		if err := s.registry.IncreaseApplicationRunningInstances(appId); err != nil {
			logrus.Errorf("Updating application got error: %s", err.Error())
		}

		app, err := s.registry.FetchApplication(appId)
		if err != nil {
			break
		}

		if app.RunningInstances == app.Instances {
			if err := s.registry.UpdateApplicationStatus(appId, "RUNNING"); err != nil {
				logrus.Errorf("Updating application got error: %s", err.Error())
			}
		}

	case mesos.TaskState_TASK_FINISHED:
		STATUS = "RESCHEDULING"
	case mesos.TaskState_TASK_FAILED:
		STATUS = "RESCHEDULING"
	case mesos.TaskState_TASK_KILLED:
		STATUS = "KILLED"
	case mesos.TaskState_TASK_LOST:
		STATUS = "RESCHEDULING"
	}

	task, err := s.registry.FetchApplicationTask(appId, taskId)
	if err != nil {
		logrus.Errorf("Fetch task %s failed: %s", taskId, err.Error())
		return
	}

	if STATUS == "RESCHEDULING" && len(task.HealthChecks) == 0 {
		msg := types.ReschedulerMsg{
			AppID:  appId,
			TaskID: taskId,
			Err:    make(chan error),
		}

		s.ReschedQueue <- msg
	}
}
