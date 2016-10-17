package scheduler

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

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
				var task types.Task

				task.ID = fmt.Sprintf("%d", time.Now().UnixNano())

				app, err := s.registry.FetchApplication(version.ID)
				if err != nil {
					return
				}
				task.Name = fmt.Sprintf("%d.%s.%s.%s", app.Instances, app.ID, app.UserId, app.ClusterId)
				app.Instances = app.Instances + 1
				if err := s.registry.UpdateApplication(app); err != nil {
					return
				}

				task.AppId = version.ID

				task.Image = version.Container.Docker.Image
				task.Network = version.Container.Docker.Network

				if version.Container.Docker.Parameters != nil {
					for _, parameter := range *version.Container.Docker.Parameters {
						task.Parameters = append(task.Parameters, &types.Parameter{
							Key:   parameter.Key,
							Value: parameter.Value,
						})
					}
				}

				if version.Container.Docker.PortMappings != nil {
					for _, portMapping := range *version.Container.Docker.PortMappings {
						task.PortMappings = append(task.PortMappings, &types.PortMappings{
							Port:     uint32(portMapping.ContainerPort),
							Protocol: portMapping.Protocol,
						})
					}
				}

				if version.Container.Docker.Privileged != nil {
					task.Privileged = version.Container.Docker.Privileged
				}

				if version.Container.Docker.ForcePullImage != nil {
					task.ForcePullImage = version.Container.Docker.ForcePullImage
				}

				task.Env = version.Env

				task.Volumes = version.Container.Volumes

				if version.Labels != nil {
					task.Labels = version.Labels
				}

				task.OfferId = offer.GetId().Value
				task.AgentId = offer.AgentId.Value
				task.AgentHostname = offer.Hostname

				if err := s.registry.RegisterTask(&task); err != nil {
					return
				}

				taskInfo := s.BuildTask(offer, resources, &task)
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
	app, err := s.registry.FetchApplication(version.ID)
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
	app, err := s.registry.FetchApplication(version.ID)
	if err != nil {
		return err
	}
	app.Status = "RUNNING"
	if err := s.registry.UpdateApplication(app); err != nil {
		return err
	}

	return nil
}
