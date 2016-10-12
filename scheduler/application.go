package scheduler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
)

func (s *Scheduler) ListApplications() ([]*types.Application, error) {
	return s.registry.ListApplications()
}

func (s *Scheduler) FetchApplication(id string) (*types.Application, error) {
	return s.registry.FetchApplication(id)
}

func (s *Scheduler) DeleteApplication(id string) error {
	return s.registry.DeleteApplication(id)
}

func (s *Scheduler) LaunchApplication(application *types.Application) error {
	if err := s.registry.RegisterApplication(application); err != nil {
		logrus.Errorf("Register application %s in consul failed: %s", application.ID, err.Error())
	}

	for i := 0; i < application.Instances; i++ {
		var task types.Task
		resources := s.BuildResources(application.Cpus, application.Mem, application.Disk)
		offer, err := s.RequestOffer(resources)
		if err != nil {
			logrus.Errorf("Request offers failed: %s", err.Error())
			return err
		}

		if offer != nil {
			task.ID = fmt.Sprintf("%d", time.Now().UnixNano())

			task.Name = application.ID

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

			resp, err := s.LaunchTask(offer, resources, &task)
			if err != nil {
				return err
			}

			if resp != nil && resp.StatusCode != http.StatusAccepted {
				return fmt.Errorf("status code %d received", resp.StatusCode)
			}
		}
	}
	return nil
}
