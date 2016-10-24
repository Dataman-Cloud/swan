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
		logrus.Infof("Task %s RUNNING", taskId)
		STATUS = "RUNNING"
		//Update application status to RUNNING
		app, err := s.registry.FetchApplication(appId)
		if err != nil {
			break
		}

		if app == nil {
			return
		}

		app.RunningInstances += 1
		if err := s.registry.UpdateApplication(app); err != nil {
			logrus.Errorf("Updating application got error: %s", err.Error())
		}

		app, err = s.registry.FetchApplication(appId)
		if err != nil {
			break
		}

		if app.RunningInstances == app.Instances {
			app.Status = "RUNNING"
		}

		if err := s.registry.UpdateApplication(app); err != nil {
			logrus.Errorf("Updating application got error: %s", err.Error())
		}

	case mesos.TaskState_TASK_FINISHED:
		STATUS = "RESCHEDULING"
	case mesos.TaskState_TASK_FAILED:
		STATUS = "RESCHEDULING"
	case mesos.TaskState_TASK_KILLED:
		STATUS = "KILLED"
	case mesos.TaskState_TASK_LOST:
	}

	task, err := s.registry.FetchApplicationTask(appId, taskId)
	if err != nil {
		logrus.WithFields(logrus.Fields{"Name": taskId, "message": status.GetMessage()}).Warn("Update received for unknown task.")
		return
	}

	if STATUS == "RESCHEDULING" && task.Status != "RESCHEDULING" {
		task.Status = STATUS
		if err := s.registry.RegisterTask(task); err != nil {
			logrus.Errorf("Register task failed: %s", err.Error())
		}

		s.HealthCheckManager.StopCheck(taskId)
		// Delete task health check
		if err := s.registry.DeleteCheck(taskId); err != nil {
			logrus.Errorf("Delete task health check %s from consul failed: %s", task.ID, err.Error())
		}

		msg := types.ReschedulerMsg{
			AppID:  appId,
			TaskID: taskId,
			Err:    make(chan error),
		}

		s.ReschedQueue <- msg
	}
}
