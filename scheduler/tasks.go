package scheduler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/types"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
)

func (s *Scheduler) BuildTask(offer *mesos.Offer, version *types.ApplicationVersion, name string) (*types.Task, error) {
	var task types.Task

	task.Name = name
	if task.Name == "" {
		app, err := s.registry.FetchApplication(version.ID)
		if err != nil {
			return nil, err
		}

		task.Name = fmt.Sprintf("%d.%s.%s.%s", app.Instances, app.ID, app.UserId, app.ClusterId)
		app.Instances = app.Instances + 1
		if err := s.registry.UpdateApplication(app); err != nil {
			return nil, err
		}
	}

	task.AppId = version.ID
	task.ID = fmt.Sprintf("%d-%s", time.Now().UnixNano(), task.Name)

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

	task.Cpus = version.Cpus
	task.Mem = version.Mem
	task.Disk = version.Disk

	task.OfferId = offer.GetId().Value
	task.AgentId = offer.AgentId.Value
	task.AgentHostname = offer.Hostname

	if version.KillPolicy != nil {
		task.KillPolicy = version.KillPolicy
	}

	if version.HealthChecks != nil {
		task.HealthChecks = version.HealthChecks
	}

	return &task, nil
}

func (s *Scheduler) BuildTaskInfo(offer *mesos.Offer, resources []*mesos.Resource, task *types.Task) *mesos.TaskInfo {
	logrus.Infof("Prepared task for launch with offer %s", *offer.GetId().Value)
	taskInfo := mesos.TaskInfo{
		Name: proto.String(task.Name),
		TaskId: &mesos.TaskID{
			Value: proto.String(task.ID),
		},
		AgentId:   offer.AgentId,
		Resources: resources,
		Command: &mesos.CommandInfo{
			Shell: proto.Bool(false),
			Value: nil,
		},
		Container: &mesos.ContainerInfo{
			Type: mesos.ContainerInfo_DOCKER.Enum(),
			Docker: &mesos.ContainerInfo_DockerInfo{
				Image: task.Image,
			},
		},
	}

	if task.Privileged != nil {
		taskInfo.Container.Docker.Privileged = task.Privileged
	}

	if task.ForcePullImage != nil {
		taskInfo.Container.Docker.ForcePullImage = task.ForcePullImage
	}

	for _, parameter := range task.Parameters {
		taskInfo.Container.Docker.Parameters = append(taskInfo.Container.Docker.Parameters, &mesos.Parameter{
			Key:   proto.String(parameter.Key),
			Value: proto.String(parameter.Value),
		})
	}

	for _, volume := range task.Volumes {
		mode := mesos.Volume_RO
		if volume.Mode == "RW" {
			mode = mesos.Volume_RW
		}
		taskInfo.Container.Volumes = append(taskInfo.Container.Volumes, &mesos.Volume{
			ContainerPath: proto.String(volume.ContainerPath),
			HostPath:      proto.String(volume.HostPath),
			Mode:          &mode,
		})
	}

	vars := make([]*mesos.Environment_Variable, 0)
	for k, v := range task.Env {
		vars = append(vars, &mesos.Environment_Variable{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}

	taskInfo.Command.Environment = &mesos.Environment{
		Variables: vars,
	}

	if task.Labels != nil {
		labels := make([]*mesos.Label, 0)
		for k, v := range *task.Labels {
			labels = append(labels, &mesos.Label{
				Key:   proto.String(k),
				Value: proto.String(v),
			})
		}

		taskInfo.Labels = &mesos.Labels{
			Labels: labels,
		}
	}

	switch task.Network {
	case "NONE":
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_NONE.Enum()
	case "HOST":
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_HOST.Enum()
	case "BRIDGE":
		ports := s.GetPorts(offer)
		for _, m := range task.PortMappings {
			hostPort := ports[s.taskLaunched]
			//ports = ports[1:]
			taskInfo.Container.Docker.PortMappings = append(taskInfo.Container.Docker.PortMappings,
				&mesos.ContainerInfo_DockerInfo_PortMapping{
					HostPort:      proto.Uint32(uint32(hostPort)),
					ContainerPort: proto.Uint32(m.Port),
					Protocol:      proto.String(m.Protocol),
				},
			)
			taskInfo.Resources = append(taskInfo.Resources, &mesos.Resource{
				Name: proto.String("ports"),
				Type: mesos.Value_RANGES.Enum(),
				Ranges: &mesos.Value_Ranges{
					Range: []*mesos.Value_Range{
						{
							Begin: proto.Uint64(uint64(hostPort)),
							End:   proto.Uint64(uint64(hostPort)),
						},
					},
				},
			})
		}
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_BRIDGE.Enum()
	default:
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_NONE.Enum()
	}

	return &taskInfo
}

// LaunchTasks lauch multiple tasks with specified offer.
func (s *Scheduler) LaunchTasks(offer *mesos.Offer, tasks []*mesos.TaskInfo) (*http.Response, error) {
	logrus.Infof("Launch %d tasks with offer %s", len(tasks), *offer.GetId().Value)
	call := &sched.Call{
		FrameworkId: s.framework.GetId(),
		Type:        sched.Call_ACCEPT.Enum(),
		Accept: &sched.Call_Accept{
			OfferIds: []*mesos.OfferID{
				offer.GetId(),
			},
			Operations: []*mesos.Offer_Operation{
				&mesos.Offer_Operation{
					Type: mesos.Offer_Operation_LAUNCH.Enum(),
					Launch: &mesos.Offer_Operation_Launch{
						TaskInfos: tasks,
					},
				},
			},
			Filters: &mesos.Filters{RefuseSeconds: proto.Float64(1)},
		},
	}

	return s.send(call)
}

func (s *Scheduler) KillTask(task *types.Task) (*http.Response, error) {
	logrus.WithFields(logrus.Fields{"ID": task.ID, "AgentId": *task.AgentId}).Warn("Kill Task")
	call := &sched.Call{
		FrameworkId: s.framework.GetId(),
		Type:        sched.Call_KILL.Enum(),
		Kill: &sched.Call_Kill{
			TaskId: &mesos.TaskID{
				Value: proto.String(task.ID),
			},
			AgentId: &mesos.AgentID{
				Value: task.AgentId,
			},
		},
	}

	if task.KillPolicy != nil {
		if task.KillPolicy.Duration != 0 {
			call.Kill.KillPolicy = &mesos.KillPolicy{
				GracePeriod: &mesos.DurationInfo{
					Nanoseconds: proto.Int64(task.KillPolicy.Duration * 1000 * 1000),
				},
			}
		}
	}

	return s.send(call)
}

// ReschedulerTask process task re-scheduler if needed.
func (s *Scheduler) ReschedulerTask() {
	for {
		select {
		case msg := <-s.ReschedQueue:
			task, err := s.registry.FetchApplicationTask(msg.AppID, msg.TaskID)
			if err != nil {
				logrus.Errorf("Rescheduler task failed: %s", err.Error())
				return
			}

			s.Status = "busy"

			resources := s.BuildResources(task.Cpus, task.Mem, task.Disk)
			offers, err := s.RequestOffers(resources)
			if err != nil {
				logrus.Errorf("Request offers failed: %s", err.Error())
				msg.Err <- err
			}

			var choosedOffer *mesos.Offer
			for _, offer := range offers {
				cpus, mem, disk := s.OfferedResources(offer)
				if cpus >= task.Cpus && mem >= task.Mem && disk >= task.Disk {
					choosedOffer = offer
					break
				}
			}

			var taskInfos []*mesos.TaskInfo
			taskInfo := s.BuildTaskInfo(choosedOffer, resources, task)
			taskInfos = append(taskInfos, taskInfo)

			resp, err := s.LaunchTasks(choosedOffer, taskInfos)
			if err != nil {
				logrus.Errorf("Launchs task failed: %s", err.Error())
				msg.Err <- err
			}

			if resp != nil && resp.StatusCode != http.StatusAccepted {
				logrus.Errorf("Launchs task failed: status code %d", resp.StatusCode)
				msg.Err <- err
			}

			msg.Err <- nil

			s.Status = "idle"

		case <-s.DoneChan:
			return
		}
	}
}
