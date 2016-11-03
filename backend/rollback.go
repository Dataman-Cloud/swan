package backend

import (
	"fmt"
	"net/http"
	"sort"

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
	if err := b.store.UpdateApplicationStatus(app.ID, "ROLLINGBACK"); err != nil {
		return err
	}

	versions, err := b.store.ListApplicationVersions(applicationId)
	if err != nil {
		return err
	}

	sort.Strings(versions)

	rollbackVer := versions[len(versions)-2]
	version, err := b.store.FetchApplicationVersion(applicationId, rollbackVer)
	if err != nil {
		return err
	}

	tasks, err := b.store.ListApplicationTasks(applicationId)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		// Stop task health check
		if b.sched.HealthCheckManager.HasCheck(task.Name) {
			b.sched.HealthCheckManager.StopCheck(task.Name)
		}

		// Delete task health check
		if err := b.store.DeleteCheck(task.Name); err != nil {
			logrus.Errorf("Delete task health check %s from consul failed: %s", task.ID, err.Error())
		}

		if _, err := b.sched.KillTask(task); err == nil {
			b.store.DeleteApplicationTask(app.ID, task.ID)
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

		if err := b.store.RegisterTask(task); err != nil {
			return err
		}

		if len(task.HealthChecks) != 0 {
			if err := b.store.RegisterCheck(task,
				*taskInfo.Container.Docker.PortMappings[0].HostPort,
				app.ID); err != nil {
			}
			for _, healthCheck := range task.HealthChecks {
				check := types.Check{
					ID:       task.Name,
					Address:  *task.AgentHostname,
					Port:     int(*taskInfo.Container.Docker.PortMappings[0].HostPort),
					TaskID:   task.Name,
					AppID:    app.ID,
					Protocol: healthCheck.Protocol,
					Interval: int(healthCheck.IntervalSeconds),
					Timeout:  int(healthCheck.TimeoutSeconds),
				}
				if healthCheck.Command != nil {
					check.Command = healthCheck.Command
				}

				if healthCheck.Path != nil {
					check.Path = *healthCheck.Path
				}

				if healthCheck.MaxConsecutiveFailures != nil {
					check.MaxFailures = *healthCheck.MaxConsecutiveFailures
				}

				b.sched.HealthCheckManager.Add(&check)
			}
		}

		b.sched.Status = "idle"
	}

	if err := b.store.UpdateApplicationStatus(app.ID, "RUNNING"); err != nil {
		return err
	}

	return nil
}
