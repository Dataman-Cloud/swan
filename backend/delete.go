package backend

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

// DeleteApplication will delete all data associated with application.
func (b *Backend) DeleteApplication(id string) error {
	tasks, err := b.store.GetTasks(id)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		// Kill task via mesos
		resp, err := b.sched.KillTask(task)
		if err != nil {
			logrus.Errorf("Kill task failed: %s", err.Error())
		}

		// Decline offer
		if resp.StatusCode == http.StatusAccepted {
			b.sched.DeclineResource(&task.OfferId)
		}

		// Stop task health check
		b.sched.HealthCheckManager.StopCheck(task.Name)

	}

	return b.store.DeleteApp(id)
}

// DeleteApplicationTasks delete all tasks belong to appcaiton but keep that application exists.
func (b *Backend) DeleteApplicationTasks(id string) error {
	tasks, err := b.store.GetTasks(id)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		// Kill task via mesos
		if _, err := b.sched.KillTask(task); err != nil {
			logrus.Errorf("Kill task failed: %s", err.Error())
		}
		//TODO delete task why not stop and delete check
	}

	return b.store.DeleteTasks(id)
}

func (b *Backend) DeleteApplicationTask(applicationId, taskId string) error {
	task, err := b.store.GetTask(applicationId, taskId)
	if err != nil {
		return err
	}

	if _, err := b.sched.KillTask(task); err != nil {
		logrus.Errorf("Kill task failed: %s", err.Error())
		return err
	}

	logrus.Infof("Stop health check for task %s", task.Name)
	b.sched.HealthCheckManager.StopCheck(task.Name)

	// Delete task health check
	if err := b.store.DeleteHealthCheck(applicationId, task.Name); err != nil {
		logrus.Errorf("Delete task health check %s from consul failed: %s", task.ID, err.Error())
	}

	return b.store.DeleteTask(applicationId, taskId)
}
