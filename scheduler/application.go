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
		resp, err := s.KillTask(*task.AgentId, task.ID)
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
		if _, err := s.KillTask(*task.AgentId, task.ID); err != nil {
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

	if _, err := s.KillTask(*task.AgentId, taskId); err != nil {
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
	app, err := s.registry.FetchApplication(applicationId)
	if err != nil {
		return err
	}
	app.Status = "SCALING"
	if err := s.registry.UpdateApplication(app); err != nil {
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

	app, err = s.registry.FetchApplication(version.ID)
	if err != nil {
		return err
	}

	if app.Instances > instances {
		tasks, err := s.registry.ListApplicationTasks(app.ID)
		if err != nil {
			return err
		}

		for _, task := range tasks {
			taskId, err := strconv.Atoi(strings.Split(task.Name, ".")[0])
			if err != nil {
				return err
			}

			if taskId+1 > instances {
				if _, err := s.KillTask(*task.AgentId, task.ID); err == nil {
					s.registry.DeleteApplicationTask(app.ID, task.ID)
				}

				app, err := s.registry.FetchApplication(version.ID)
				if err != nil {
					return err
				}
				app.Instances = app.Instances - 1
				if err := s.registry.UpdateApplication(app); err != nil {
					return err
				}

			}
		}
	}

	if app.Instances < instances {
		version.Instances = instances - app.Instances
		s.LaunchApplication(version)
	}

	// Update application status to RUNNING
	app, err = s.registry.FetchApplication(version.ID)
	if err != nil {
		return err
	}
	app.Status = "RUNNING"
	if err := s.registry.UpdateApplication(app); err != nil {
		return err
	}

	return nil
}

// UpdateApplication is used for application rolling-update.
func (s *Scheduler) UpdateApplication(applicationId string, instances int, version *types.ApplicationVersion) error {
	// Update application status to UPDATING
	app, err := s.registry.FetchApplication(version.ID)
	if err != nil {
		return err
	}
	app.Status = "UPDATING"
	if err := s.registry.UpdateApplication(app); err != nil {
		return err
	}

	tasks, err := s.registry.ListApplicationTasks(app.ID)

	begin, end := app.InstanceUpdated, app.InstanceUpdated+instances
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
				if _, err := s.KillTask(*task.AgentId, task.ID); err == nil {
					s.registry.DeleteApplicationTask(app.ID, task.ID)
				}

				// reduce application instance count.
				app.Instances = app.Instances - 1
				if err := s.registry.UpdateApplication(app); err != nil {
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

				// increase application updated instance count.
				app.InstanceUpdated += 1
				if err := s.registry.UpdateApplication(app); err != nil {
					return err
				}

				// increase application instance count.
				app.Instances += 1
				if err := s.registry.UpdateApplication(app); err != nil {
					return err
				}
				s.Status = "idle"

			}

		}

	}

	// Rest application updated instance count to zero.
	if app.InstanceUpdated == app.Instances {
		app.InstanceUpdated = 0
		if err := s.registry.UpdateApplication(app); err != nil {
			return err
		}
	}

	// Update application status to RUNNING
	app.Status = "RUNNING"
	if err := s.registry.UpdateApplication(app); err != nil {
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

	app.Status = "ROLLINGBACK"
	if err := s.registry.UpdateApplication(app); err != nil {
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
		if _, err := s.KillTask(*task.AgentId, task.ID); err == nil {
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

		s.Status = "idle"
	}
	// Update application status to UPDATING
	app.Status = "RUNNING"
	if err := s.registry.UpdateApplication(app); err != nil {
		return err
	}

	return nil
}
