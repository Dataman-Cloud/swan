package backend

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/types"

	"github.com/Sirupsen/logrus"
)

// UpdateApplication is used for application rolling-update.
func (b *Backend) UpdateApplication(applicationId string, instances int, version *types.Version) error {
	logrus.Infof("Updating application %s", applicationId)
	app, err := b.FetchApplication(applicationId)
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
	if err := b.store.UpdateAppStatus(applicationId, "UPDATING"); err != nil {
		logrus.Errorf("Setting application %s status to UPDATING for rolling-update failed: %s", applicationId, err.Error())
		return err
	}

	tasks, err := b.store.GetTasks(applicationId)
	if err != nil {
		logrus.Errorf("List application %s tasks failed: %s", applicationId, err.Error())
		return err
	}

	begin, end := int(app.UpdatedInstances), int(app.UpdatedInstances)+instances
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
							if maxFailover >= int(version.UpdatePolicy.MaxFailovers) {
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

	if instances == -1 {
		begin, end = 0, len(tasks)
	}

	go func() error {
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
					if err := b.store.DeleteHealthCheck(app.ID, task.Name); err != nil {
						logrus.Errorf("Delete task health check %s from consul failed: %s", task.ID, err.Error())
					}

					if _, err := b.sched.KillTask(task); err == nil {
						b.store.DeleteTask(app.ID, task.ID)
					}

					//Reduce application running instance count.
					if err := b.store.AddAppRunningInstance(applicationId, -1); err != nil {
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

					if err := b.store.PutTask(applicationId, task); err != nil {
						return err
					}

					// Pause updating
					time.Sleep(time.Duration(version.UpdatePolicy.UpdateDelay) * time.Second)

					if len(task.HealthChecks) != 0 {
						if err := b.store.PutHealthcheck(task,
							*taskInfo.Container.Docker.PortMappings[0].HostPort,
							version.ID); err != nil {
						}
						for _, healthCheck := range task.HealthChecks {
							check := types.Check{
								ID:       task.Name,
								Address:  task.AgentHostname,
								Port:     int64(*taskInfo.Container.Docker.PortMappings[0].HostPort),
								TaskID:   task.Name,
								AppID:    version.ID,
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

					if err := b.store.AddAppUpdatedInstance(app.ID, 1); err != nil {
						return err
					}

					app, err := b.store.GetApp(applicationId)
					if err != nil {
						return err
					}

					// Rest application updated instance count to zero.
					if app.UpdatedInstances == app.Instances {
						if err := b.store.UpdateAppUpdatedInstance(app.ID, 0); err != nil {
							return err
						}
						// Update application status to RUNNING
						if err := b.store.UpdateAppStatus(app.ID, "RUNNING"); err != nil {
							return err
						}

						logrus.Infof("Updating application %s finished", app.ID)
					}

				}
			}
		}
		return nil
	}()

	return nil
}
