package scheduler

import (
	"github.com/Sirupsen/logrus"
	"net/http"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
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

	// task, err := s.registry.Fetch(ID)
	// if err != nil {
	// 	logrus.WithFields(logrus.Fields{"ID": ID, "message": status.GetMessage()}).Warn("Update received for unknown task.")
	// 	return
	// }

	// task.State = &state
	// if err := s.registry.Update(ID, task); err != nil {
	// 	logrus.WithFields(logrus.Fields{"ID": ID, "message": status.GetMessage(), "error": err}).Error("Update task state in registry")
	// }

	switch state {
	case mesos.TaskState_TASK_STAGING:
		logrus.WithFields(logrus.Fields{"ID": ID, "message": status.GetMessage()}).Info("Task was registered.")
	case mesos.TaskState_TASK_STARTING:
		logrus.WithFields(logrus.Fields{"ID": ID, "message": status.GetMessage()}).Info("Task is starting.")
	case mesos.TaskState_TASK_RUNNING:
		logrus.WithFields(logrus.Fields{"ID": ID, "message": status.GetMessage()}).Info("Task is running.")
		// if !status.GetHealthy() {
		// 	logrus.WithFields(
		// 		logrus.Fields{"ID": ID, "message": status.GetMessage()},
		// 	).Info("Task is RUNNING, but service seems Failed")
		// } else {
		// 	logrus.WithFields(logrus.Fields{"ID": ID, "status": status.GetState()}).Info("Task is RUNNING")
		// }

	case mesos.TaskState_TASK_FINISHED:
		logrus.WithFields(logrus.Fields{"ID": ID, "status": status.GetState(), "message": status.GetMessage()}).Info("Task is finished.")
		//s.reScheduler(ID)
	case mesos.TaskState_TASK_FAILED:
		logrus.WithFields(logrus.Fields{"ID": ID, "status": status.GetState(), "message": status.GetMessage()}).Warn("Task has failed.")
		//s.reScheduler(ID)
	case mesos.TaskState_TASK_KILLED:
		logrus.WithFields(logrus.Fields{"ID": ID, "status": status.GetState(), "message": status.GetMessage()}).Warn("Task was killed.")
		//s.reScheduler(ID)
	case mesos.TaskState_TASK_LOST:
		logrus.WithFields(logrus.Fields{"ID": ID, "status": status.GetState(), "message": status.GetMessage()}).Warn("Task was lost.")
		//s.reScheduler(ID)
	}
}

// func (s *Scheduler) reScheduler(ID string) {
// 	logrus.WithFields(logrus.Fields{"ID": ID}).Info("Try to re-scheduler")
// 	// task, _ := s.registry.Fetch(ID)
// 	resources := s.BuildResources(task.Cpus, task.Mem, task.Disk)
// 	offer, _ := s.RequestOffer(resources)
// 	s.LaunchTask(offer, resources, task)
// }
