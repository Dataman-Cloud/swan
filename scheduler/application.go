package scheduler

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
)

func (s *Scheduler) ListApplications() ([]*types.Application, error) {
	return s.registry.ListApplications()
}

func (s *Scheduler) FetchApplication(id string) (*types.Application, error) {
	return s.registry.FetchApplication(id)
}

// DeleteApplication will delete all data associated with application.
func (s *Scheduler) DeleteApplication(id string) error {
	tasks, err := s.registry.ListApplicationTasks(id)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		// Kill task via mesos
		resp, err := s.KillTask(task)
		if err != nil {
			logrus.Errorf("Kill task failed: %s", err.Error())
		}

		// Decline offer
		if resp.StatusCode == http.StatusAccepted {
			s.DeclineResource(task.OfferId)
		}

		// Delete task from consul
		if err := s.registry.DeleteApplicationTask(id, task.ID); err != nil {
			logrus.Errorf("Delete task %s from consul failed: %s", task.ID, err.Error())
		}

		// Delete task health check
		if err := s.registry.DeleteCheck(task.Name); err != nil {
			logrus.Errorf("Delete task health check %s from consul failed: %s", task.ID, err.Error())
		}

		// Stop task health check
		s.HealthCheckManager.StopCheck(task.Name)

	}

	return s.registry.DeleteApplication(id)
}

func (s *Scheduler) ListApplicationTasks(id string) ([]*types.Task, error) {
	return s.registry.ListApplicationTasks(id)
}

// DeleteApplicationTasks delete all tasks belong to appcaiton but keep that application exists.
func (s *Scheduler) DeleteApplicationTasks(id string) error {
	tasks, err := s.registry.ListApplicationTasks(id)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		// Kill task via mesos
		if _, err := s.KillTask(task); err != nil {
			logrus.Errorf("Kill task failed: %s", err.Error())
		}

		// Delete task from consul
		if err := s.registry.DeleteApplicationTask(id, task.ID); err != nil {
			logrus.Errorf("Delete task %s from consul failed: %s", task.ID, err.Error())
		}
	}

	return nil
}

func (s *Scheduler) DeleteApplicationTask(applicationId, taskId string) error {
	task, err := s.registry.FetchApplicationTask(applicationId, taskId)
	if err != nil {
		return err
	}

	if err := s.registry.DeleteApplicationTask(applicationId, taskId); err != nil {
		return err
	}

	if _, err := s.KillTask(task); err != nil {
		logrus.Errorf("Kill task failed: %s", err.Error())
		return err
	}

	return nil
}

// RegisterApplication register application in consul.
func (s *Scheduler) RegisterApplication(application *types.Application) error {
	return s.registry.RegisterApplication(application)
}

// RegisterApplicationVersion register application version in consul.
func (s *Scheduler) RegisterApplicationVersion(version *types.ApplicationVersion) error {
	return s.registry.RegisterApplicationVersion(version)
}

func (s *Scheduler) LaunchApplication(version *types.ApplicationVersion) error {
	s.taskLaunched = 0

	// Set scheduler's status to busy for accepting resource.
	s.Status = "busy"

	go func() {
		resources := s.BuildResources(version.Cpus, version.Mem, version.Disk)
		offers, err := s.RequestOffers(resources)
		if err != nil {
			logrus.Errorf("Request offers failed: %s", err.Error())
		}

		for _, offer := range offers {
			cpus, mem, disk := s.OfferedResources(offer)
			var tasks []*mesos.TaskInfo
			for s.taskLaunched < version.Instances &&
				cpus >= version.Cpus &&
				mem >= version.Mem &&
				disk >= version.Disk {
				task, err := s.BuildTask(offer, version, "")
				if err != nil {
					logrus.Errorf("Build task failed: %s", err.Error())
					return
				}

				if err := s.registry.RegisterTask(task); err != nil {
					return
				}

				taskInfo := s.BuildTaskInfo(offer, resources, task)
				tasks = append(tasks, taskInfo)

				if len(task.HealthChecks) != 0 {
					if err := s.registry.RegisterCheck(task,
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

						s.HealthCheckManager.Add(&check)
					}
				}

				s.taskLaunched++
				cpus -= version.Cpus
				mem -= version.Mem
				disk -= version.Disk
			}

			resp, err := s.LaunchTasks(offer, tasks)
			if err != nil {
				logrus.Errorf("Launchs task failed: %s", err.Error())
			}

			if resp != nil && resp.StatusCode != http.StatusAccepted {
				logrus.Errorf("status code %d received", resp.StatusCode)
			}
		}

		// Set scheduler's status back to idle after launch applicaiton.
		s.Status = "idle"
	}()

	return nil
}

// ListApplicationVersions is used to list all versions for application from consul specified by application id.
func (s *Scheduler) ListApplicationVersions(applicationId string) ([]string, error) {
	return s.registry.ListApplicationVersions(applicationId)
}

// FetchApplicationVersion is used to fetch specified version from consul by version id and application id.
func (s *Scheduler) FetchApplicationVersion(applicationId, versionId string) (*types.ApplicationVersion, error) {
	return s.registry.FetchApplicationVersion(applicationId, versionId)
}

// ScaleApplication is used to scale application instances.
func (s *Scheduler) ScaleApplication(applicationId string, instances int) error {
	// Update application status to SCALING
	if err := s.registry.UpdateApplication(applicationId, "status", "SCALING"); err != nil {
		logrus.Errorf("Updating application status to SCALING failed: %s", err.Error())
		return err
	}

	versions, err := s.registry.ListApplicationVersions(applicationId)
	if err != nil {
		return err
	}

	sort.Strings(versions)

	newestVersion := versions[len(versions)-1]
	version, err := s.registry.FetchApplicationVersion(applicationId, newestVersion)
	if err != nil {
		return err
	}

	app, err := s.registry.FetchApplication(version.ID)
	if err != nil {
		return err
	}

	if app.Instances > instances {
		tasks, err := s.registry.ListApplicationTasks(app.ID)
		if err != nil {
			return err
		}

		for _, task := range tasks {
			taskIndex, err := strconv.Atoi(strings.Split(task.Name, ".")[0])
			if err != nil {
				return err
			}

			if taskIndex+1 > instances {
				if _, err := s.KillTask(task); err == nil {
					s.registry.DeleteApplicationTask(app.ID, task.ID)
				}

				// reduce application tasks count
				if err := s.registry.ReduceApplicationInstances(app.ID); err != nil {
					logrus.Errorf("Updating application %s instances count failed: %s", app.ID, err.Error())
					return err
				}

				logrus.Infof("Remove health check for task %s", task.Name)

				s.HealthCheckManager.StopCheck(task.Name)

				if err := s.registry.DeleteCheck(task.Name); err != nil {
					logrus.Errorf("Remove health check for %s failed: %s", task.Name, err.Error())
					return err
				}

				if err := s.registry.DeleteApplicationTask(app.ID, task.Name); err != nil {
					logrus.Errorf("Delete task %s failed: %s", task.Name, err.Error())
				}

			}
		}
	}

	if app.Instances < instances {
		for i := 0; i < instances-app.Instances; i++ {
			s.Status = "busy"

			resources := s.BuildResources(version.Cpus, version.Mem, version.Disk)
			offers, err := s.RequestOffers(resources)
			if err != nil {
				logrus.Errorf("Request offers failed: %s for rescheduling", err.Error())
			}

			var choosedOffer *mesos.Offer
			for _, offer := range offers {
				cpus, mem, disk := s.OfferedResources(offer)
				if cpus >= version.Cpus && mem >= version.Mem && disk >= version.Disk {
					choosedOffer = offer
					break
				}
			}

			name := fmt.Sprintf("%d.%s.%s.%s", app.Instances+i, app.ID, app.UserId, app.ClusterId)

			task, err := s.BuildTask(choosedOffer, version, name)
			if err != nil {
				logrus.Errorf("Build task failed: %s", err.Error())
				return err
			}

			var taskInfos []*mesos.TaskInfo
			taskInfo := s.BuildTaskInfo(choosedOffer, resources, task)
			taskInfos = append(taskInfos, taskInfo)

			resp, err := s.LaunchTasks(choosedOffer, taskInfos)
			if err != nil {
				logrus.Errorf("Launchs task failed: %s", err.Error())
				return err
			}

			if resp != nil && resp.StatusCode != http.StatusAccepted {
				return fmt.Errorf("status code %d received", resp.StatusCode)
			}

			if err := s.registry.RegisterTask(task); err != nil {
				return err
			}

			if len(version.HealthChecks) != 0 {
				if err := s.registry.RegisterCheck(task,
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

					s.HealthCheckManager.Add(&check)
				}
			}

			// Increase application task count
			if err := s.registry.IncreaseApplicationInstances(version.ID); err != nil {
				logrus.Errorf("Updating application %s instance count failed: %s", version.ID, err.Error())
				return err
			}

			s.Status = "idle"
		}
	}

	// Update application status to RUNNING
	// app, err = s.registry.FetchApplication(version.ID)
	// if err != nil {
	// 	return err
	// }
	// app.Status = "RUNNING"
	// if err := s.registry.UpdateApplication(app); err != nil {
	// 	return err
	// }
	if err := s.registry.UpdateApplication(version.ID, "status", "RUNNING"); err != nil {
		logrus.Errorf("Updating application %s status to RUNNING failed: %s", version.ID, err.Error())
		return err
	}

	return nil
}

// UpdateApplication is used for application rolling-update.
func (s *Scheduler) UpdateApplication(applicationId string, instances int, version *types.ApplicationVersion) error {
	// Update application status to UPDATING
	// app.Status = "UPDATING"
	// if err := s.registry.UpdateApplication(app); err != nil {
	// 	return err
	// }
	if err := s.registry.UpdateApplication(applicationId, "status", "UPDATING"); err != nil {
		logrus.Errorf("Setting application %s status to UPDATING for rolling-update failed: %s", applicationId, err.Error())
		return err
	}

	app, err := s.registry.FetchApplication(applicationId)
	if err != nil {
		return err
	}

	tasks, err := s.registry.ListApplicationTasks(applicationId)

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
				if _, err := s.KillTask(task); err == nil {
					s.registry.DeleteApplicationTask(app.ID, task.ID)
				}

				// reduce application instance count.
				// app.Instances = app.Instances - 1
				// if err := s.registry.UpdateApplication(app); err != nil {
				// 	return err
				// }
				if err := s.registry.UpdateApplication(applicationId, "instance", "-1"); err != nil {
					return err
				}

				s.Status = "busy"

				resources := s.BuildResources(version.Cpus, version.Mem, version.Disk)
				offers, err := s.RequestOffers(resources)
				if err != nil {
					logrus.Errorf("Request offers failed: %s", err.Error())
				}

				var choosedOffer *mesos.Offer
				for _, offer := range offers {
					cpus, mem, disk := s.OfferedResources(offer)
					if cpus >= version.Cpus && mem >= version.Mem && disk >= version.Disk {
						choosedOffer = offer
						break
					}
				}

				task, err := s.BuildTask(choosedOffer, version, task.Name)
				if err != nil {
					logrus.Errorf("Build task failed: %s", err.Error())
					return err
				}

				if err := s.registry.RegisterTask(task); err != nil {
					return err
				}

				var taskInfos []*mesos.TaskInfo
				taskInfo := s.BuildTaskInfo(choosedOffer, resources, task)
				taskInfos = append(taskInfos, taskInfo)

				resp, err := s.LaunchTasks(choosedOffer, taskInfos)
				if err != nil {
					logrus.Errorf("Launchs task failed: %s", err.Error())
					return err
				}

				if resp != nil && resp.StatusCode != http.StatusAccepted {
					return fmt.Errorf("status code %d received", resp.StatusCode)
				}

				if len(task.HealthChecks) != 0 {
					if err := s.registry.RegisterCheck(task,
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

						s.HealthCheckManager.Add(&check)
					}
				}

				// increase application updated instance count.
				//app.UpdatedInstances += 1
				//if err := s.registry.UpdateApplication(app); err != nil {
				//	return err
				//}
				if err := s.registry.IncreaseApplicationUpdatedInstances(app.ID); err != nil {
					return err
				}

				// increase application instance count.
				// app.Instances += 1
				// if err := s.registry.UpdateApplication(app); err != nil {
				// 	return err
				// }
				if err := s.registry.IncreaseApplicationInstances(app.ID); err != nil {
					return err
				}

				s.Status = "idle"

			}

		}

	}

	// Rest application updated instance count to zero.
	if app.UpdatedInstances == app.Instances {
		if err := s.registry.ResetApplicationUpdatedInstances(app.ID); err != nil {
			return err
		}
		// app.UpdatedInstances = 0
		// if err := s.registry.UpdateApplication(app); err != nil {
		// 	return err
		// }
	}

	// Update application status to RUNNING
	// app.Status = "RUNNING"
	// if err := s.registry.UpdateApplication(app); err != nil {
	// 	return err
	// }
	if err := s.registry.UpdateApplicationStatus(app.ID, "RUNNING"); err != nil {
		return err
	}

	return nil
}

// RollbackApplication rollback application to previous version.
func (s *Scheduler) RollbackApplication(applicationId string) error {
	// Update application status to UPDATING
	app, err := s.registry.FetchApplication(applicationId)
	if err != nil {
		return err
	}

	if app == nil {
		logrus.Errorf("Application %s not found for rollback", applicationId)
		return errors.New("Application not found")
	}

	// app.Status = "ROLLINGBACK"
	// if err := s.registry.UpdateApplication(app); err != nil {
	// 	return err
	// }
	if err := s.registry.UpdateApplicationStatus(app.ID, "ROLLINGBACK"); err != nil {
		return err
	}

	versions, err := s.registry.ListApplicationVersions(applicationId)
	if err != nil {
		return err
	}

	sort.Strings(versions)

	rollbackVer := versions[len(versions)-2]
	version, err := s.registry.FetchApplicationVersion(applicationId, rollbackVer)
	if err != nil {
		return err
	}

	tasks, err := s.registry.ListApplicationTasks(applicationId)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if _, err := s.KillTask(task); err == nil {
			s.registry.DeleteApplicationTask(app.ID, task.ID)
		}

		s.Status = "busy"

		resources := s.BuildResources(version.Cpus, version.Mem, version.Disk)
		offers, err := s.RequestOffers(resources)
		if err != nil {
			logrus.Errorf("Request offers failed: %s", err.Error())
		}

		var choosedOffer *mesos.Offer
		for _, offer := range offers {
			cpus, mem, disk := s.OfferedResources(offer)
			if cpus >= version.Cpus && mem >= version.Mem && disk >= version.Disk {
				choosedOffer = offer
				break
			}
		}

		task, err := s.BuildTask(choosedOffer, version, task.Name)
		if err != nil {
			logrus.Errorf("Build task failed: %s", err.Error())
			return err
		}

		var taskInfos []*mesos.TaskInfo
		taskInfo := s.BuildTaskInfo(choosedOffer, resources, task)
		taskInfos = append(taskInfos, taskInfo)

		resp, err := s.LaunchTasks(choosedOffer, taskInfos)
		if err != nil {
			logrus.Errorf("Launchs task failed: %s", err.Error())
			return err
		}

		if resp != nil && resp.StatusCode != http.StatusAccepted {
			return fmt.Errorf("status code %d received", resp.StatusCode)
		}

		if err := s.registry.RegisterTask(task); err != nil {
			return err
		}

		if len(task.HealthChecks) != 0 {
			if err := s.registry.RegisterCheck(task,
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

				s.HealthCheckManager.Add(&check)
			}
		}

		s.Status = "idle"
	}
	// Update application status to UPDATING
	//app.Status = "RUNNING"
	//if err := s.registry.UpdateApplication(app); err != nil {
	//	return err
	//}
	if err := s.registry.UpdateApplicationStatus(app.ID, "RUNNING"); err != nil {
		return err
	}

	return nil
}
