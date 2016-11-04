package backend

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/types"

	"github.com/Sirupsen/logrus"
)

// ScaleApplication is used to scale application instances.
func (b *Backend) ScaleApplication(applicationId string, instances int) error {
	app, err := b.store.GetApp(applicationId)
	if err != nil {
		return err
	}

	if app.Status != "RUNNING" {
		return errors.New("Operation Not Allowed")
	}

	// Update application status to SCALING
	if err := b.store.UpdateAppStatus(applicationId, "SCALING"); err != nil {
		logrus.Errorf("Updating application status to SCALING failed: %s", err.Error())
		return err
	}

	versions, err := b.store.GetAndSortVersions(applicationId)
	if err != nil {
		return err
	}

	if len(versions) < 1 {
		errors.New("app versions was empty")
	}

	version := versions[len(versions)-1]

	go func() error {
		if int(app.Instances) > instances {
			tasks, err := b.store.GetTasks(app.ID)
			if err != nil {
				return err
			}

			for _, task := range tasks {
				taskIndex, err := strconv.Atoi(strings.Split(task.Name, ".")[0])
				if err != nil {
					return err
				}

				if taskIndex+1 > instances {
					b.sched.HealthCheckManager.StopCheck(task.Name)

					if err := b.store.DeleteHealthCheck(applicationId, task.Name); err != nil {
						logrus.Errorf("Remove health check for %s failed: %s", task.Name, err.Error())
						return err
					}

					if _, err := b.sched.KillTask(task); err == nil {
						b.store.DeleteTask(app.ID, task.ID)
					}

					// reduce application tasks count
					if err := b.store.AddAppInstance(app.ID, -1); err != nil {
						logrus.Errorf("Updating application %s instances count failed: %s", app.ID, err.Error())
						return err
					}

					logrus.Infof("Remove health check for task %s", task.Name)

					if err := b.store.DeleteTask(app.ID, task.Name); err != nil {
						logrus.Errorf("Delete task %s failed: %s", task.Name, err.Error())
					}

				}
			}
		}

		if int(app.Instances) < instances {
			for i := 0; i < instances-int(app.Instances); i++ {
				b.sched.Status = "busy"

				resources := b.sched.BuildResources(version.Cpus, version.Mem, version.Disk)
				offers, err := b.sched.RequestOffers(resources)
				if err != nil {
					logrus.Errorf("Request offers failed: %s for rescheduling", err.Error())
				}

				var choosedOffer *mesos.Offer
				for _, offer := range offers {
					cpus, mem, disk := b.sched.OfferedResources(offer)
					if cpus >= version.Cpus && mem >= version.Mem && disk >= version.Disk {
						choosedOffer = offer
						break
					}
				}
				name := fmt.Sprintf("%d.%s.%s.%s", int(app.Instances)+i, app.ID, app.UserId, app.ClusterId)

				task, err := b.sched.BuildTask(choosedOffer, version, name)
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

				if len(version.HealthChecks) != 0 {
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

				// Increase application task count
				if err := b.store.AddAppInstance(version.ID, 1); err != nil {
					logrus.Errorf("Updating application %s instance count failed: %s", version.ID, err.Error())
					return err
				}

				b.sched.Status = "idle"
			}
		}

		// Update application status to RUNNING
		if err := b.store.UpdateAppStatus(version.ID, "RUNNING"); err != nil {
			logrus.Errorf("Updating application %s status to RUNNING failed: %s", version.ID, err.Error())
			return err
		}

		return nil
	}()

	return nil
}
