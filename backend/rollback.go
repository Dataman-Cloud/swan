package backend

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
)

// RollbackApplication rollback application to previous version.
func (b *Backend) RollbackApplication(applicationId string) error {
	logrus.Infof("Rollback application %s", applicationId)
	app, err := b.store.GetApp(applicationId)
	if err != nil {
		return err
	}

	// Update application status to ROLLINGBACK
	if err := b.store.UpdateAppStatus(app.ID, "ROLLINGBACK"); err != nil {
		return err
	}

	versions, err := b.store.GetAndSortVersions(applicationId)
	if err != nil {
		return err
	}

	if len(versions) < 2 {
		return errors.New("not have history version to rollback")
	}
	version := versions[len(versions)-2]

	tasks, err := b.store.GetTasks(applicationId)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		// Stop task health check
		if b.sched.HealthCheckManager.HasCheck(task.Name) {
			b.sched.HealthCheckManager.StopCheck(task.Name)
		}

		// Delete task health check
		if err := b.store.DeleteHealthCheck(applicationId, task.Name); err != nil {
			logrus.Errorf("Delete task health check %s from consul failed: %s", task.ID, err.Error())
		}

		if _, err := b.sched.KillTask(task); err == nil {
			b.store.DeleteTask(app.ID, task.ID)
		}

		b.sched.Status = "busy"

		resources := b.sched.BuildResources(version.Cpus, version.Mem, version.Disk)
		offers, err := b.sched.RequestOffers(resources)
		if err != nil {
			logrus.Errorf("Request offers failed: %s", err.Error())
		}

		var choosedOffer *mesos.Offer
		for _, offer := range offers {
			cpus, mem, disk := b.sched.OfferedResources(offer)
			if cpus >= version.Cpus && mem >= version.Mem && disk >= version.Disk {
				choosedOffer = offer
				break
			}
		}

		task, err := b.sched.BuildTask(choosedOffer, version, task.Name)
		if err != nil {
			logrus.Errorf("Build task failed: %s", err.Error())
			return err
		}

		var taskInfos []*mesos.TaskInfo
		taskInfo := b.sched.BuildTaskInfo(choosedOffer, resources, task)
		taskInfos = append(taskInfos, taskInfo)

		resp, err := b.sched.LaunchTasks(choosedOffer, taskInfos)
		if err != nil {
			logrus.Errorf("Launchs task failed: %s", err.Error())
			return err
		}

		if resp != nil && resp.StatusCode != http.StatusAccepted {
			return fmt.Errorf("status code %d received", resp.StatusCode)
		}

		if err := b.store.PutTask(applicationId, task); err != nil {
			return err
		}

		if len(task.HealthChecks) != 0 {
			if err := b.store.PutHealthcheck(task,
				*taskInfo.Container.Docker.PortMappings[0].HostPort,
				app.ID); err != nil {
			}
			for _, healthCheck := range task.HealthChecks {
				check := types.Check{
					ID:       task.Name,
					Address:  task.AgentHostname,
					Port:     int64(*taskInfo.Container.Docker.PortMappings[0].HostPort),
					TaskID:   task.Name,
					AppID:    app.ID,
					Protocol: healthCheck.Protocol,
					Interval: int64(healthCheck.IntervalSeconds),
					Timeout:  int64(healthCheck.TimeoutSeconds),
				}
				if healthCheck.Command != nil {
					check.Command = healthCheck.Command
				}

				check.Path = healthCheck.Path
				check.MaxFailures = healthCheck.MaxConsecutiveFailures

				b.sched.HealthCheckManager.Add(&check)
			}
		}

		b.sched.Status = "idle"
	}

	if err := b.store.UpdateAppStatus(app.ID, "RUNNING"); err != nil {
		return err
	}

	return nil
}
