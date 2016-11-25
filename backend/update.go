package backend

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"sort"
	"time"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/types"

	"github.com/Sirupsen/logrus"
)

// UpdateApplication is used for application rolling-update.
func (b *Backend) UpdateApplication(appId string, instances int, version *types.Version) error {
	logrus.Infof("Updating application %s", appId)
	app, err := b.store.FetchApplication(appId)
	if err != nil {
		return err
	}

	if app == nil {
		return errors.New("Application not found")
	}

	if app.Status != "RUNNING" {
		return errors.New("Operation Not Allowed")
	}

	// Update application status to UPDATING
	if err := b.store.UpdateApplicationStatus(appId, "UPDATING"); err != nil {
		logrus.Errorf("Setting application %s status to UPDATING for rolling-update failed: %s", appId, err.Error())
		return err
	}

	tasks, err := b.store.ListTasks(appId)
	if err != nil {
		logrus.Errorf("List application %s tasks failed: %s", appId, err.Error())
		return err
	}

	sort.Sort(TaskSorter(tasks))

	begin, end := app.UpdatedInstances, app.UpdatedInstances+instances
	if instances == -1 {
		begin, end = 0, len(tasks)
	}

	if end > app.Instances {
		end = app.Instances
	}

	go func() error {
		if err := b.doUpdate(tasks[begin:end], version); err != nil {
			logrus.Errorf("Update application %s failed, rollback to previous version.", appId)
			return b.RollbackApplication(appId)
		}

		app, err = b.store.FetchApplication(appId)
		if err != nil {
			return err
		}

		// Rest application updated instance count to zero.
		if app.UpdatedInstances == app.Instances {
			if err := b.store.ResetApplicationUpdatedInstances(app.ID); err != nil {
				return err
			}

		}
		// Update application status to RUNNING
		if err := b.store.UpdateApplicationStatus(app.ID, "RUNNING"); err != nil {
			return err
		}

		logrus.Infof("Updating application %s finished", app.ID)

		return nil
	}()

	return nil
}

// doUpdate update application instances one by one.
func (b *Backend) doUpdate(tasks []*types.Task, version *types.Version) error {
	for _, task := range tasks {
		// Stop task health check
		b.sched.HealthCheckManager.StopCheck(task.Name)

		// Delete task health check
		if err := b.store.DeleteCheck(task.Name); err != nil {
			logrus.Errorf("Delete task health check %s from db failed: %s", task.ID, err.Error())
		}

		if _, err := b.sched.KillTask(task); err == nil {
			b.store.DeleteTask(task.ID)
		}

		//Reduce application running instance count.
		if err := b.store.ReduceApplicationRunningInstances(task.AppId); err != nil {
			return err
		}

		b.sched.Status = "busy"

		logrus.Infof("Launch task %s with new version", task.Name)

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
		}

		if resp != nil && resp.StatusCode != http.StatusAccepted {
			logrus.Errorf("status code %d received", resp.StatusCode)
		}

		b.sched.Status = "idle"

		if err := b.store.SaveTask(task); err != nil {
			return err
		}

		if err := b.doCheck(
			fmt.Sprintf("%s:%d",
				*choosedOffer.Hostname,
				int(*taskInfo.Container.Docker.PortMappings[0].HostPort)),
			version.UpdatePolicy); err != nil {
			return err
		}

		if len(task.HealthChecks) != 0 {
			if err := b.store.SaveCheck(task,
				*taskInfo.Container.Docker.PortMappings[0].HostPort,
				version.ID); err != nil {
			}
			for _, healthCheck := range task.HealthChecks {
				check := types.Check{
					ID:       task.Name,
					Address:  *task.AgentHostname,
					Port:     int(*taskInfo.Container.Docker.PortMappings[0].HostPort),
					TaskID:   task.Name,
					AppID:    version.ID,
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

				if healthCheck.ConsecutiveFailures != 0 {
					check.MaxFailures = int(healthCheck.ConsecutiveFailures)
				}

				b.sched.HealthCheckManager.Add(&check)

			}
		}
		//increase application running instance count.
		if err := b.store.IncreaseApplicationRunningInstances(task.AppId); err != nil {
			return err
		}

		if err := b.store.IncreaseApplicationUpdatedInstances(task.AppId); err != nil {
			return err
		}
	}

	return nil
}

func (b *Backend) doCheck(addr string, update *types.UpdatePolicy) error {
	ticker := time.NewTicker(time.Duration(2) * time.Second)

	quit := time.After(time.Duration(update.UpdateDelay) * time.Second)

	failureTimes := 0

	for {
		if failureTimes >= update.MaxFailovers {
			return fmt.Errorf("Service Update Failed")
		}

		select {
		case <-ticker.C:
			_, err := net.ResolveTCPAddr("tcp", addr)
			if err != nil {
				logrus.Errorf("Resolve tcp addr failed: %s", err.Error())
				return err
			}

			conn, err := net.DialTimeout("tcp",
				addr,
				time.Duration(2*time.Second))
			if err != nil {
				failureTimes++
			}
			if conn != nil {
				conn.Close()
			}
		case <-quit:
			ticker.Stop()
			return nil
		}
	}
}
