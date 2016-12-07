package state

import (
	"strings"

	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
)

const (
	TASK_STATE_RUNNING = "task_running"
	TASK_STATE_FAIL    = "task_failed"
	TASK_STATE_FINISH  = "task_finish"
)

type Task struct {
	App     *App
	Version *types.Version
	Slot    *Slot

	State  string
	Stdout string
	Stderr string

	OfferId       string
	AgentId       string
	Ip            string
	AgentHostName string

	Reason string
}

func NewTask(app *App, version *types.Version, slot *Slot) *Task {
	task := &Task{
		App:     app,
		Version: version,
		Slot:    slot,
	}

	return task
}

func (task *Task) PrepareTaskInfo(offer *mesos.Offer) *mesos.TaskInfo {
	logrus.Infof("Prepared task %s for launch with offer %s", task.Slot.Id, *offer.GetId().Value)

	task.OfferId = *offer.GetId().Value
	task.Slot.OfferId = *offer.GetId().Value

	task.AgentId = *offer.GetAgentId().Value
	task.Slot.AgentId = *offer.GetAgentId().Value

	task.AgentHostName = offer.GetHostname()
	task.Slot.AgentHostName = offer.GetHostname()

	taskInfo := mesos.TaskInfo{
		Name: proto.String(task.Slot.Id),
		TaskId: &mesos.TaskID{
			Value: proto.String(task.Slot.Id),
		},
		AgentId:   offer.AgentId,
		Resources: task.Slot.Resources(),
		Command: &mesos.CommandInfo{
			Shell: proto.Bool(false),
			Value: nil,
		},
		Container: &mesos.ContainerInfo{
			Type: mesos.ContainerInfo_DOCKER.Enum(),
			Docker: &mesos.ContainerInfo_DockerInfo{
				Image: &task.Slot.Version.Container.Docker.Image,
			},
		},
	}

	taskInfo.Container.Docker.Privileged = &task.Slot.Version.Container.Docker.Privileged
	taskInfo.Container.Docker.ForcePullImage = &task.Slot.Version.Container.Docker.ForcePullImage

	for _, parameter := range task.Slot.Version.Container.Docker.Parameters {
		taskInfo.Container.Docker.Parameters = append(taskInfo.Container.Docker.Parameters, &mesos.Parameter{
			Key:   proto.String(parameter.Key),
			Value: proto.String(parameter.Value),
		})
	}

	for _, volume := range task.Slot.Version.Container.Volumes {
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
	for k, v := range task.Slot.Version.Env {
		vars = append(vars, &mesos.Environment_Variable{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}

	taskInfo.Command.Environment = &mesos.Environment{
		Variables: vars,
	}

	uris := make([]*mesos.CommandInfo_URI, 0)
	for _, v := range task.Slot.Version.Uris {
		uris = append(uris, &mesos.CommandInfo_URI{
			Value: proto.String(v),
		})
	}

	if len(uris) > 0 {
		taskInfo.Command.Uris = uris
	}

	if task.Slot.Version.Labels != nil {
		labels := make([]*mesos.Label, 0)
		for k, v := range task.Slot.Version.Labels {
			labels = append(labels, &mesos.Label{
				Key:   proto.String(k),
				Value: proto.String(v),
			})
		}

		taskInfo.Labels = &mesos.Labels{
			Labels: labels,
		}
	}

	switch task.Slot.Version.Container.Docker.Network {
	case "NONE":
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_NONE.Enum()
	case "HOST":
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_HOST.Enum()
	case "BRIDGE":
		//ports := GetPorts(offer)
		//if len(ports) == 0 {
		//logrus.Errorf("No ports resource defined")
		//break
		//}
		for _, m := range task.Slot.Version.Container.Docker.PortMappings {
			hostPort := 31000
			taskInfo.Container.Docker.PortMappings = append(taskInfo.Container.Docker.PortMappings,
				&mesos.ContainerInfo_DockerInfo_PortMapping{
					HostPort:      proto.Uint32(uint32(hostPort)),
					ContainerPort: proto.Uint32(uint32(m.ContainerPort)),
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

	// setup task health check
	if len(task.Slot.Version.HealthChecks) > 0 {
		for _, healthCheck := range task.Slot.Version.HealthChecks {
			if healthCheck.PortIndex < 0 || int(healthCheck.PortIndex) > len(taskInfo.Container.Docker.PortMappings) {
				healthCheck.PortIndex = 0
			}

			hostPort := proto.Uint32(0)

			for _, portMapping := range taskInfo.Container.Docker.PortMappings {
				if portMapping.ContainerPort == proto.Uint32(uint32(healthCheck.Port)) {
					hostPort = portMapping.HostPort
				}
			}

			if healthCheck.PortName != "" {
				for _, portMapping := range task.Slot.Version.Container.Docker.PortMappings {
					if portMapping.Name == healthCheck.PortName {
						containerPort := portMapping.ContainerPort
						for _, portMapping := range taskInfo.Container.Docker.PortMappings {
							if uint32(containerPort) == *portMapping.ContainerPort {
								hostPort = portMapping.HostPort
							}
						}
					}
				}
			}

			if *hostPort == 0 {
				hostPort = taskInfo.Container.Docker.PortMappings[healthCheck.PortIndex].HostPort
			}

			protocol := strings.ToLower(healthCheck.Protocol)
			if protocol == "http" {
				taskInfo.HealthCheck = &mesos.HealthCheck{
					Type: mesos.HealthCheck_HTTP.Enum(),
					Http: &mesos.HealthCheck_HTTPCheckInfo{
						Scheme:   proto.String("http"),
						Port:     hostPort,
						Path:     &healthCheck.Path,
						Statuses: []uint32{uint32(200)},
					},
				}
			}

			if protocol == "tcp" {
				taskInfo.HealthCheck = &mesos.HealthCheck{
					Type: mesos.HealthCheck_TCP.Enum(),
					Tcp: &mesos.HealthCheck_TCPCheckInfo{
						Port: hostPort,
					},
				}
			}

			taskInfo.HealthCheck.IntervalSeconds = proto.Float64(healthCheck.IntervalSeconds)
			taskInfo.HealthCheck.TimeoutSeconds = proto.Float64(healthCheck.TimeoutSeconds)
			taskInfo.HealthCheck.ConsecutiveFailures = proto.Uint32(healthCheck.MaxConsecutiveFailures)
			taskInfo.HealthCheck.GracePeriodSeconds = proto.Float64(healthCheck.GracePeriodSeconds)
		}
	}

	return &taskInfo
}

func createScalarResource(name string, value float64) *mesos.Resource {
	return &mesos.Resource{
		Name:   &name,
		Type:   mesos.Value_SCALAR.Enum(),
		Scalar: &mesos.Value_Scalar{Value: &value},
	}
}

func createRangeResource(name string, begin, end uint64) *mesos.Resource {
	return &mesos.Resource{
		Name: &name,
		Type: mesos.Value_RANGES.Enum(),
		Ranges: &mesos.Value_Ranges{
			Range: []*mesos.Value_Range{
				{
					Begin: proto.Uint64(begin),
					End:   proto.Uint64(end),
				},
			},
		},
	}
}

func (task *Task) Kill() {
	logrus.Infof("Kill task %s", task.Slot.Id)
	call := &sched.Call{
		FrameworkId: task.App.MesosConnector.Framework.GetId(),
		Type:        sched.Call_KILL.Enum(),
		Kill: &sched.Call_Kill{
			TaskId: &mesos.TaskID{
				Value: proto.String(task.Slot.Id),
			},
			AgentId: &mesos.AgentID{
				Value: &task.AgentId,
			},
		},
	}

	if task.Version.KillPolicy != nil {
		if task.Version.KillPolicy.Duration != 0 {
			call.Kill.KillPolicy = &mesos.KillPolicy{
				GracePeriod: &mesos.DurationInfo{
					Nanoseconds: proto.Int64(task.Version.KillPolicy.Duration * 1000 * 1000),
				},
			}
		}
	}

	task.App.MesosConnector.MesosCallChan <- call
}
