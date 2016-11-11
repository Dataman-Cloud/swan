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
			logrus.Errorf("Acknowledge call returned unexpected status: %d", resp.StatusCode)
			return
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
		if err := s.store.UpdateTaskStatus(taskId, "STAGING"); err != nil {
			logrus.Errorf("updating task %s status to STAGING failed", taskId)
		}
	case mesos.TaskState_TASK_STARTING:
		STATUS = "STARTING"
		if err := s.store.UpdateTaskStatus(taskId, "STARTING"); err != nil {
			logrus.Errorf("updating task %s status to STARTING failed", taskId)
		}
	case mesos.TaskState_TASK_RUNNING:
		STATUS = "RUNNING"
		if err := s.store.IncreaseApplicationRunningInstances(appId); err != nil {
			logrus.Errorf("Updating application got error: %s", err.Error())
		}

		if err := s.store.UpdateTaskStatus(taskId, "RUNNING"); err != nil {
			logrus.Errorf("updating task %s status to RUNNING failed: %s", taskId, err.Error())
			return
		}

		app, err := s.store.FetchApplication(appId)
		if err != nil {
			break
		}

		if app.RunningInstances == app.Instances && app.Status != "UPDATING" {
			if err := s.store.UpdateApplicationStatus(appId, "RUNNING"); err != nil {
				logrus.Errorf("Updating application got error: %s", err.Error())
			}
		}

	case mesos.TaskState_TASK_FINISHED:
		logrus.Infof("Task Finished, message: %s", status.GetMessage())
		STATUS = "RESCHEDULING"
	case mesos.TaskState_TASK_FAILED:
		logrus.Infof("Task Failed, message: %s", status.GetMessage())
		STATUS = "RESCHEDULING"
	case mesos.TaskState_TASK_KILLED:
		logrus.Infof("Task Killed, message: %s", status.GetMessage())
		STATUS = "KILLED"
	case mesos.TaskState_TASK_LOST:
		logrus.Infof("Task Lost, message: %s", status.GetMessage())
		STATUS = "RESCHEDULING"
	}

	task, err := s.store.FetchTask(taskId)
	if err != nil {
		logrus.Errorf("Fetch task %s failed: %s", taskId, err.Error())
		return
	}

	app, err := s.store.FetchApplication(appId)
	if err != nil {
		logrus.Errorf("Fetch application %s failed: %s", appId, err.Error())
		return
	}

	if STATUS == "RESCHEDULING" &&
		len(task.HealthChecks) == 0 &&
		task.Status != "RESCHEDULING" &&
		app.Status != "UPDATING" &&
		app.Status != "ROLLINGBACK" {
		if err := s.store.UpdateTaskStatus(taskId, "RESCHEDULING"); err != nil {
			logrus.Errorf("updating task status to RESCHEDULING failed: %s", taskId)
		}

		s.Status = "busy"

		resources := s.BuildResources(task.Cpus, task.Mem, task.Disk)
		offers, err := s.RequestOffers(resources)
		if err != nil {
			logrus.Errorf("Request offers failed: %s for rescheduling", err.Error())
			return
		}

		var choosedOffer *mesos.Offer
		for _, offer := range offers {
			cpus, mem, disk := s.OfferedResources(offer)
			if cpus >= task.Cpus && mem >= task.Mem && disk >= task.Disk {
				choosedOffer = offer
				break
			}
		}

		var taskInfos []*mesos.TaskInfo
		taskInfo := s.BuildTaskInfo(choosedOffer, resources, task)
		taskInfos = append(taskInfos, taskInfo)

		resp, err := s.LaunchTasks(choosedOffer, taskInfos)
		if err != nil {
			logrus.Errorf("Launchs task failed: %s for rescheduling", err.Error())
		}

		if resp != nil && resp.StatusCode != http.StatusAccepted {
			logrus.Errorf("Launchs task failed: status code %d for rescheduling", resp.StatusCode)
		}
	}
}
