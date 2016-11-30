package backend

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
)

// RollbackApplication rollback application to previous version.
func (b *Backend) RollbackApplication(appId string) error {
	logrus.Infof("Rollback application %s", appId)
	app, err := b.store.FetchApplication(appId)
	if err != nil {
		return err
	}

	if app == nil {
		logrus.Errorf("Application %s not found for rollback", appId)
		return errors.New("Application not found")
	}

	// Update application status to ROLLINGBACK
	if err := b.store.UpdateApplicationStatus(app.ID, "ROLLINGBACK"); err != nil {
		return err
	}

	versions, err := b.store.ListVersions(appId)
	if err != nil {
		return err
	}

	sort.Strings(versions)

	rollbackVer := versions[len(versions)-2]
	version, err := b.store.FetchVersion(rollbackVer)
	if err != nil {
		return err
	}

	tasks, err := b.store.ListTasks(appId)
	if err != nil {
		return err
	}

	go func() error {
		if err := b.doRollback(tasks, version); err != nil {
			logrus.Errorf("Rollback application failed: %s", appId)
			if err := b.store.UpdateApplicationStatus(app.ID, "ROLLBACK-FAILED"); err != nil {
				return err
			}
			return err
		}

		if err := b.store.UpdateApplicationStatus(app.ID, "RUNNING"); err != nil {
			return err
		}

		return nil
	}()

	return nil
}

func (b *Backend) doRollback(tasks []*types.Task, version *types.Version) error {
	for _, task := range tasks {
		// Stop task health check
		if b.sched.HealthCheckManager.HasCheck(task.Name) {
			b.sched.HealthCheckManager.StopCheck(task.Name)
		}

		// Delete task health check
		if err := b.store.DeleteCheck(task.Name); err != nil {
			logrus.Errorf("Delete task health check %s from db failed: %s", task.ID, err.Error())
		}

		if _, err := b.sched.KillTask(task); err == nil {
			b.store.DeleteTask(task.AppId, task.ID)
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

		if err := b.store.SaveTask(task); err != nil {
			return err
		}

		// TODO: (pwzgorilla) clear unuse code
		//if len(task.HealthChecks) != 0 {
		//	if err := b.store.SaveCheck(task,
		//		*taskInfo.Container.Docker.PortMappings[0].HostPort,
		//		task.AppId); err != nil {
		//	}
		//	for _, healthCheck := range task.HealthChecks {
		//		check := types.Check{
		//			ID:       task.Name,
		//			Address:  *task.AgentHostname,
		//			Port:     int(*taskInfo.Container.Docker.PortMappings[0].HostPort),
		//			TaskID:   task.Name,
		//			AppID:    task.AppId,
		//			Protocol: healthCheck.Protocol,
		//			Interval: int(healthCheck.IntervalSeconds),
		//			Timeout:  int(healthCheck.TimeoutSeconds),
		//		}
		//		if healthCheck.Command != nil {
		//			check.Command = healthCheck.Command
		//		}

		//		if healthCheck.Path != nil {
		//			check.Path = *healthCheck.Path
		//		}

		//		if healthCheck.ConsecutiveFailures != 0 {
		//			check.MaxFailures = int(healthCheck.ConsecutiveFailures)
		//		}

		//		b.sched.HealthCheckManager.Add(&check)
		//	}
		//}

		b.sched.Status = "idle"
	}

	return nil
}
