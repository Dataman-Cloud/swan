package scheduler

import (
	"net/http"
	"strings"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
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
		logrus.WithFields(logrus.Fields{"Name": taskId, "message": status.GetMessage()}).Info("Task was registered.")
		STATUS = "STAGING"
	case mesos.TaskState_TASK_STARTING:
		logrus.WithFields(logrus.Fields{"Name": taskId, "message": status.GetMessage()}).Info("Task is starting.")
		STATUS = "STARTING"
	case mesos.TaskState_TASK_RUNNING:
		logrus.WithFields(logrus.Fields{"Name": taskId, "message": status.GetMessage()}).Info("Task is running.")

		STATUS = "RUNNING"

		//Update application status to RUNNING
		app, err := s.registry.FetchApplication(appId)
		if err != nil {
			break
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
		logrus.WithFields(logrus.Fields{"Name": taskId, "status": status.GetState(), "message": status.GetMessage()}).Info("Task is finished.")
		STATUS = "FINISHED"
	case mesos.TaskState_TASK_FAILED:
		logrus.WithFields(logrus.Fields{"Name": taskId, "status": status.GetState(), "message": status.GetMessage()}).Warn("Task has failed.")
		STATUS = "FAILED"
	case mesos.TaskState_TASK_KILLED:
		logrus.WithFields(logrus.Fields{"Name": taskId, "status": status.GetState(), "message": status.GetMessage()}).Warn("Task was killed.")
		STATUS = "KILLED"
	case mesos.TaskState_TASK_LOST:
		logrus.WithFields(logrus.Fields{"Name": taskId, "status": status.GetState(), "message": status.GetMessage()}).Warn("Task was lost.")
		STATUS = "LOST"
	}

	task, err := s.registry.FetchApplicationTask(appId, taskId)
	if err != nil {
		logrus.WithFields(logrus.Fields{"Name": taskId, "message": status.GetMessage()}).Warn("Update received for unknown task.")
		return
	}

	task.Status = STATUS
	if err := s.registry.RegisterTask(task); err != nil {
		logrus.WithFields(logrus.Fields{"Name": taskId, "message": status.GetMessage(), "error": err}).Error("Update task state in registry")
	}

}
