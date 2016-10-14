package scheduler

import (
	"fmt"
	"net/http"
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

func (s *Scheduler) LaunchApplication(application *types.Application) error {
	if err := s.registry.RegisterApplication(application); err != nil {
		logrus.Errorf("Register application %s in consul failed: %s", application.ID, err.Error())
	}

	s.taskLaunched = 0

	go func() {
		resources := s.BuildResources(application.Cpus, application.Mem, application.Disk)
		offers, err := s.RequestOffers(resources)
		if err != nil {
			logrus.Errorf("Request offers failed: %s", err.Error())
		}

		for _, offer := range offers {
			cpus, mem, disk := s.OfferedResources(offer)
			var tasks []*mesos.TaskInfo
			for s.taskLaunched < application.Instances &&
				cpus >= application.Cpus &&
				mem >= application.Mem &&
				disk >= application.Disk {
				var task types.Task

				task.ID = fmt.Sprintf("%d", time.Now().UnixNano())
				task.Name = fmt.Sprintf("%d.%s", application.AutoIncrement, application.ID)
				application.AutoIncrement++
				task.AppId = application.ID

				task.Image = application.Container.Docker.Image
				task.Network = application.Container.Docker.Network

				if application.Container.Docker.Parameters != nil {
					for _, parameter := range *application.Container.Docker.Parameters {
						task.Parameters = append(task.Parameters, &types.Parameter{
							Key:   parameter.Key,
							Value: parameter.Value,
						})
					}
				}

				if application.Container.Docker.PortMappings != nil {
					for _, portMapping := range *application.Container.Docker.PortMappings {
						task.PortMappings = append(task.PortMappings, &types.PortMappings{
							Port:     uint32(portMapping.ContainerPort),
							Protocol: portMapping.Protocol,
						})
					}
				}

				if application.Container.Docker.Privileged != nil {
					task.Privileged = application.Container.Docker.Privileged
				}

				if application.Container.Docker.ForcePullImage != nil {
					task.ForcePullImage = application.Container.Docker.ForcePullImage
				}

				task.Env = application.Env

				task.Volumes = application.Container.Volumes

				if application.Labels != nil {
					task.Labels = application.Labels
				}

				task.OfferId = offer.GetId().Value
				task.AgentId = offer.AgentId.Value
				task.AgentHostname = offer.Hostname

				taskInfo := s.BuildTask(offer, resources, &task)
				tasks = append(tasks, taskInfo)
				s.taskLaunched++
				cpus -= application.Cpus
				mem -= application.Mem
				disk -= application.Disk
			}

			resp, err := s.LaunchTasks(offer, tasks)
			if err != nil {
				logrus.Errorf("Launchs task failed: %s", err.Error())
			}

			if resp != nil && resp.StatusCode != http.StatusAccepted {
				logrus.Errorf("status code %d received", resp.StatusCode)
			}
		}
	}()

	return nil
}
