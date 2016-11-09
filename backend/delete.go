package backend

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

// DeleteApplication will delete all data associated with application.
func (b *Backend) DeleteApplication(appId string) error {
	tasks, err := b.store.ListTasks(appId)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		// Stop task health check
		b.sched.HealthCheckManager.StopCheck(task.Name)

		// Kill task via mesos
		resp, err := b.sched.KillTask(task)
		if err != nil {
			logrus.Errorf("Kill task failed: %s", err.Error())
			break
		}

		// Decline offer
		if resp.StatusCode == http.StatusAccepted {
			b.sched.DeclineResource(task.OfferId)
		}

		// Delete task from db
		if err := b.store.DeleteTask(task.Name); err != nil {
			logrus.Errorf("Delete task %s from db failed: %s", task.ID, err.Error())
		}

		// Delete task health check
		if err := b.store.DeleteCheck(task.Name); err != nil {
			logrus.Errorf("Delete task health check %s from db failed: %s", task.ID, err.Error())
		}

	}

	versions, err := b.store.ListVersions(appId)
	if err != nil {
		return err
	}

	for _, version := range versions {
		if err := b.store.DeleteVersion(version); err != nil {
			return err
		}
	}

	return b.store.DeleteApplication(appId)
}

// DeleteApplicationTasks delete all tasks belong to appcaiton but keep that application exists.
func (b *Backend) DeleteApplicationTasks(id string) error {
	tasks, err := b.store.ListTasks(id)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		// Kill task via mesos
		if _, err := b.sched.KillTask(task); err != nil {
			logrus.Errorf("Kill task failed: %s", err.Error())
		}

		// Delete task from db
		if err := b.store.DeleteTask(task.ID); err != nil {
			logrus.Errorf("Delete task %s from db failed: %s", task.ID, err.Error())
		}
	}

	return nil
}

func (b *Backend) DeleteApplicationTask(applicationId, taskId string) error {
	task, err := b.store.FetchTask(taskId)
	if err != nil {
		return err
	}

	if err := b.store.DeleteTask(taskId); err != nil {
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
		logrus.Errorf("Delete task health check %s from db failed: %s", task.ID, err.Error())
	}

	return nil
}
