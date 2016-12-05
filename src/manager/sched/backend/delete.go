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
		// Kill task via mesos
		resp, err := b.sched.KillTask(task)
		if err != nil {
			logrus.Errorf("Kill task failed: %s", err.Error())
			break
		}

		// Decline offer
		if resp.StatusCode == http.StatusAccepted {
			b.sched.DeclineResource(&task.OfferId)
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
		if err := b.store.DeleteTask(id, task.ID); err != nil {
			logrus.Errorf("Delete task %s from db failed: %s", task.ID, err.Error())
		}
	}

	return nil
}

func (b *Backend) DeleteApplicationTask(applicationId, taskId string) error {
	task, err := b.store.FetchTask(applicationId, taskId)
	if err != nil {
		return err
	}

	if err := b.store.DeleteTask(applicationId, taskId); err != nil {
		return err
	}

	if _, err := b.sched.KillTask(task); err != nil {
		logrus.Errorf("Kill task failed: %s", err.Error())
		return err
	}

	return nil
}
