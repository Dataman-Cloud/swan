package backend

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/types"

	"github.com/Sirupsen/logrus"
)

// UpdateApplication is used for application rolling-update.
func (b *Backend) UpdateApplication(applicationId string, instances int, version *types.ApplicationVersion) error {
	logrus.Infof("Updating application %s", applicationId)
	app, err := b.store.FetchApplication(applicationId)
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
	if err := b.store.UpdateApplicationStatus(applicationId, "UPDATING"); err != nil {
		logrus.Errorf("Setting application %s status to UPDATING for rolling-update failed: %s", applicationId, err.Error())
		return err
	}

	tasks, err := b.store.ListApplicationTasks(applicationId)
	if err != nil {
		logrus.Errorf("List application %s tasks failed: %s", applicationId, err.Error())
		return err
	}

	maxFailover := 0
	quit := make(chan struct{})

	taskNames := make([]string, 0)
	for _, task := range tasks {
		taskNames = append(taskNames, task.Name)
	}

	go func() {
		for {
			select {
			case event := <-b.sched.GetEvent(sched.Event_UPDATE):
				status := event.GetUpdate().GetStatus()
				ID := status.TaskId.GetValue()
				taskName := strings.Split(ID, "-")[1]
				for _, name := range taskNames {
					if name == taskName {
						switch status.GetState() {
						case mesos.TaskState_TASK_FAILED,
							mesos.TaskState_TASK_FINISHED,
							mesos.TaskState_TASK_LOST:
							maxFailover++
							if maxFailover >= version.UpdatePolicy.MaxFailovers {
								switch version.UpdatePolicy.Action {
								case "rollback", "ROLLBACK":
									logrus.Errorf("MaxFailovers exceeded, Rollback")
									go func() {
										b.RollbackApplication(version.ID)
									}()
									close(quit)
								case "stop", "STOP":
									logrus.Errorf("MaxFailovers exceeded, Stop Updating")
									close(quit)
								}
							}
						}
					}
				}
			case <-quit:
				return
			}
		}
	}()

	begin, end := app.UpdatedInstances, app.UpdatedInstances+instances
	if instances == -1 {
		begin, end = 0, len(tasks)+1
	}

	for i := begin; i < end; i++ {
		for _, task := range tasks {
			taskId, err := strconv.Atoi(strings.Split(task.Name, ".")[0])
			if err != nil {
				return err
			}

			if taskId == i {
				// Stop task health check
				b.sched.HealthCheckManager.StopCheck(task.Name)

				// Delete task health check
				if err := b.store.DeleteCheck(task.Name); err != nil {
					logrus.Errorf("Delete task health check %s from consul failed: %s", task.ID, err.Error())
				}

				if _, err := b.sched.KillTask(task); err == nil {
					b.store.DeleteApplicationTask(app.ID, task.ID)
				}

				// Reduce application instance count.
				if err := b.store.UpdateApplication(applicationId, "instance", "-1"); err != nil {
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

				if err := b.sched.SyncLaunchTask(choosedOffer, taskInfos); err != nil {
					for i := 0; i < version.UpdatePolicy.MaxRetries; i++ {
						logrus.Infof("Launch task %s failed, retry.", task.Name)
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

						if err := b.sched.SyncLaunchTask(choosedOffer, taskInfos); err == nil {
							break
						}

						if i == version.UpdatePolicy.MaxRetries-1 {
							logrus.Errorf(`Launch task %s failed after retry 3 times,
							finished updating and rollback to previous version.`, task.Name)
							return b.RollbackApplication(version.ID)
						}
					}
				}

				if err := b.store.RegisterTask(task); err != nil {
					return err
				}

				// Pause updating
				time.Sleep(time.Duration(version.UpdatePolicy.UpdateDelay) * time.Second)

				if len(task.HealthChecks) != 0 {
					if err := b.store.RegisterCheck(task,
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

						if healthCheck.MaxConsecutiveFailures != nil {
							check.MaxFailures = *healthCheck.MaxConsecutiveFailures
						}

						b.sched.HealthCheckManager.Add(&check)

					}
				}

				if err := b.store.IncreaseApplicationUpdatedInstances(app.ID); err != nil {
					return err
				}

				if err := b.store.IncreaseApplicationInstances(app.ID); err != nil {
					return err
				}

				app, err := b.store.FetchApplication(applicationId)
				if err != nil {
					return err
				}

				// Rest application updated instance count to zero.
				if app.UpdatedInstances == app.Instances {
					if err := b.store.ResetApplicationUpdatedInstances(app.ID); err != nil {
						return err
					}
					// Update application status to RUNNING
					if err := b.store.UpdateApplicationStatus(app.ID, "RUNNING"); err != nil {
						return err
					}

					logrus.Infof("Updating application %s finished", app.ID)
				}

				b.sched.Status = "idle"

				close(quit)
			}
		}
	}

	return nil
}
