package backend

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

// DeleteApplication will delete all data associated with application.
func (b *Backend) DeleteApplication(id string) error {
	tasks, err := b.store.ListApplicationTasks(id)
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
			b.sched.DeclineResource(task.OfferId)
		}

		// Delete task from consul
		if err := b.store.DeleteApplicationTask(id, task.ID); err != nil {
			logrus.Errorf("Delete task %s from consul failed: %s", task.ID, err.Error())
		}

		// Delete task health check
		if err := b.store.DeleteCheck(task.Name); err != nil {
			logrus.Errorf("Delete task health check %s from consul failed: %s", task.ID, err.Error())
		}

		// Stop task health check
		b.sched.HealthCheckManager.StopCheck(task.Name)

	}

	return b.store.DeleteApplication(id)
}

// DeleteApplicationTasks delete all tasks belong to appcaiton but keep that application exists.
func (b *Backend) DeleteApplicationTasks(id string) error {
	tasks, err := b.store.ListApplicationTasks(id)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		// Kill task via mesos
		if _, err := b.sched.KillTask(task); err != nil {
			logrus.Errorf("Kill task failed: %s", err.Error())
		}

		// Delete task from consul
		if err := b.store.DeleteApplicationTask(id, task.ID); err != nil {
			logrus.Errorf("Delete task %s from consul failed: %s", task.ID, err.Error())
		}
	}

	return nil
}

func (b *Backend) DeleteApplicationTask(applicationId, taskId string) error {
	task, err := b.store.FetchApplicationTask(applicationId, taskId)
	if err != nil {
		return err
	}

	if err := b.store.DeleteApplicationTask(applicationId, taskId); err != nil {
		return err
	}

	if _, err := b.sched.KillTask(task); err != nil {
		logrus.Errorf("Kill task failed: %s", err.Error())
		return err
	}

	logrus.Infof("Stop health check for task %s", task.Name)
	b.sched.HealthCheckManager.StopCheck(task.Name)

	// Delete task health check
	if err := b.store.DeleteCheck(task.Name); err != nil {
		logrus.Errorf("Delete task health check %s from consul failed: %s", task.ID, err.Error())
	}

	return nil
}
