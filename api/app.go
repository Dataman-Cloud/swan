package api

import (
	"encoding/json"
	"fmt"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	"net/http"
	"time"
)

func (r *Router) appCreate(w http.ResponseWriter, req *http.Request) error {
	var application types.Application

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&application); err != nil {
		return err
	}

	for i := 0; i < application.Instances; i++ {
		var task types.Task
		resources := r.sched.BuildResources(application.Cpus, application.Mem, application.Disk)
		offer, err := r.sched.RequestOffer(resources)
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

			resp, err := r.sched.LaunchTask(offer, resources, &task)
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
