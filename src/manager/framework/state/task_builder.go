package state

import (
	"fmt"
	"strings"

	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/golang/protobuf/proto"
)

type TaskBuilder struct {
	task      *Task
	taskInfo  *mesos.TaskInfo
	HostPorts []uint64
}

func NewTaskBuilder(task *Task) *TaskBuilder {
	builder := &TaskBuilder{
		task:      task,
		taskInfo:  &mesos.TaskInfo{},
		HostPorts: make([]uint64, 0),
	}

	builder.taskInfo.Labels = &mesos.Labels{Labels: make([]*mesos.Label, 0)}

	return builder
}

func (builder *TaskBuilder) GetTaskInfo() *mesos.TaskInfo {
	return builder.taskInfo
}

func (builder *TaskBuilder) SetName(name string) *TaskBuilder {
	builder.taskInfo.Name = proto.String(name)

	return builder
}

func (builder *TaskBuilder) SetTaskId(taskId string) *TaskBuilder {
	builder.taskInfo.TaskId = &mesos.TaskID{
		Value: proto.String(taskId),
	}

	return builder
}

func (builder *TaskBuilder) SetAgentId(agentId string) *TaskBuilder {
	builder.taskInfo.AgentId = &mesos.AgentID{
		Value: proto.String(agentId),
	}

	return builder
}

func (builder *TaskBuilder) SetResources(resources []*mesos.Resource) *TaskBuilder {
	builder.taskInfo.Resources = resources

	return builder
}

func (builder *TaskBuilder) SetCommand(cmd string, args []string) *TaskBuilder {
	if len(cmd) > 0 {
		builder.taskInfo.Command = &mesos.CommandInfo{
			Shell:     proto.Bool(true),
			Arguments: args,
			Value:     proto.String(cmd),
		}
	} else {
		builder.taskInfo.Command = &mesos.CommandInfo{
			Shell: proto.Bool(false),
		}
	}

	return builder
}

func (builder *TaskBuilder) SetContainerType(containerType string) *TaskBuilder {
	if containerType == "docker" {

		builder.taskInfo.Container = &mesos.ContainerInfo{
			Type:   mesos.ContainerInfo_DOCKER.Enum(),
			Docker: &mesos.ContainerInfo_DockerInfo{},
		}
	}

	return builder
}

func (builder *TaskBuilder) SetContainerDockerImage(image string) *TaskBuilder {
	builder.taskInfo.Container.Docker.Image = proto.String(image)

	return builder
}

func (builder *TaskBuilder) SetContainerDockerPrivileged(privileged bool) *TaskBuilder {
	builder.taskInfo.Container.Docker.Privileged = proto.Bool(privileged)

	return builder
}

func (builder *TaskBuilder) SetContainerDockerForcePullImage(force bool) *TaskBuilder {
	builder.taskInfo.Container.Docker.ForcePullImage = proto.Bool(force)

	return builder
}

func (builder *TaskBuilder) AppendContainerDockerParameters(parameters []*types.Parameter) *TaskBuilder {
	for _, parameter := range parameters {
		builder.taskInfo.Container.Docker.Parameters = append(builder.taskInfo.Container.Docker.Parameters, &mesos.Parameter{
			Key:   proto.String(parameter.Key),
			Value: proto.String(parameter.Value),
		})
	}

	return builder
}

func (builder *TaskBuilder) AppendContainerDockerVolumes(volumes []*types.Volume) *TaskBuilder {
	for _, volume := range volumes {
		mode := mesos.Volume_RO
		if strings.ToLower(volume.Mode) == "rw" {
			mode = mesos.Volume_RW
		}

		builder.taskInfo.Container.Volumes = append(builder.taskInfo.Container.Volumes, &mesos.Volume{
			ContainerPath: proto.String(volume.ContainerPath),
			HostPath:      proto.String(volume.HostPath),
			Mode:          &mode,
		})
	}

	return builder
}

func (builder *TaskBuilder) AppendContainerDockerEnvironments(envs map[string]string) *TaskBuilder {
	vars := make([]*mesos.Environment_Variable, 0)

	if builder.taskInfo.Command.Environment != nil {
		vars = builder.taskInfo.Command.Environment.Variables
	}

	for k, v := range envs {
		vars = append(vars, &mesos.Environment_Variable{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}

	builder.taskInfo.Command.Environment = &mesos.Environment{
		Variables: vars,
	}

	return builder
}

func (builder *TaskBuilder) SetURIs(uriList []string) *TaskBuilder {
	uris := make([]*mesos.CommandInfo_URI, 0)
	for _, v := range uriList {
		uris = append(uris, &mesos.CommandInfo_URI{
			Value: proto.String(v),
		})
	}

	if len(uris) > 0 {
		builder.taskInfo.Command.Uris = uris
	}

	return builder
}

func (builder *TaskBuilder) AppendTaskInfoLabels(labelMap map[string]string) *TaskBuilder {
	for k, v := range labelMap {
		builder.taskInfo.Labels.Labels = append(builder.taskInfo.Labels.Labels, &mesos.Label{
			Key:   proto.String(k),
			Value: proto.String(v),
		})
	}

	return builder
}

func (builder *TaskBuilder) SetNetwork(network string, portsAvailable []uint64) *TaskBuilder {
	builder.HostPorts = make([]uint64, 0) // clear this array on every loop
	portsRelatedEnvs := make(map[string]string)
	switch strings.ToLower(network) {
	case "none":
		builder.taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_NONE.Enum()
	case "host":
		for index, m := range builder.task.Slot.Version.Container.Docker.PortMappings {
			hostPort := uint64(m.HostPort)
			if m.HostPort == 0 { // random port when host port is 0
				hostPort = portsAvailable[index]
				builder.taskInfo.Resources = append(builder.taskInfo.Resources, &mesos.Resource{
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
			builder.HostPorts = append(builder.HostPorts, hostPort)
			portsRelatedEnvs[fmt.Sprintf("SWAN_HOST_PORT_%s", strings.ToUpper(m.Name))] = fmt.Sprintf("%d", hostPort)
		}
		builder.taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_HOST.Enum()
	case "bridge":
		for index, m := range builder.task.Slot.Version.Container.Docker.PortMappings {
			hostPort := portsAvailable[index]
			builder.HostPorts = append(builder.HostPorts, hostPort)
			builder.taskInfo.Container.Docker.PortMappings = append(builder.taskInfo.Container.Docker.PortMappings,
				&mesos.ContainerInfo_DockerInfo_PortMapping{
					HostPort:      proto.Uint32(uint32(hostPort)),
					ContainerPort: proto.Uint32(uint32(m.ContainerPort)),
					Protocol:      proto.String(m.Protocol),
				},
			)

			portsRelatedEnvs[fmt.Sprintf("SWAN_HOST_PORT_%s", strings.ToUpper(m.Name))] = fmt.Sprintf("%d", hostPort)
			portsRelatedEnvs[fmt.Sprintf("SWAN_CONTAINER_PORT_%s", strings.ToUpper(m.Name))] = fmt.Sprintf("%d", m.ContainerPort)

			builder.taskInfo.Resources = append(builder.taskInfo.Resources, &mesos.Resource{
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
		builder.taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_BRIDGE.Enum()

	default:
		builder.taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_USER.Enum()
		builder.taskInfo.Container.NetworkInfos = append(builder.taskInfo.Container.NetworkInfos, &mesos.NetworkInfo{
			Name: proto.String(network),
		})
	}

	builder.AppendContainerDockerEnvironments(portsRelatedEnvs)

	return builder
}

func (builder *TaskBuilder) SetHealthCheck(healthCheck *types.HealthCheck) *TaskBuilder {
	protocol := strings.ToLower(healthCheck.Protocol)
	if protocol == "cmd" {
		builder.taskInfo.HealthCheck = &mesos.HealthCheck{
			Type: mesos.HealthCheck_COMMAND.Enum(),
			Command: &mesos.CommandInfo{
				Value: &healthCheck.Value,
			},
		}
	} else {
		var namespacePort int32
		for _, portMapping := range builder.task.Slot.Version.Container.Docker.PortMappings {
			if portMapping.Name == healthCheck.PortName {
				if strings.ToLower(builder.task.Slot.Version.Container.Docker.Network) == "host" {
					namespacePort = portMapping.HostPort
				} else if strings.ToLower(builder.task.Slot.Version.Container.Docker.Network) == "bridge" {
					namespacePort = portMapping.ContainerPort
				} else { // not support, shortcut
					return builder
				}
			}
		}

		if protocol == "http" {
			builder.taskInfo.HealthCheck = &mesos.HealthCheck{
				Type: mesos.HealthCheck_HTTP.Enum(),
				Http: &mesos.HealthCheck_HTTPCheckInfo{
					Scheme:   proto.String(protocol),
					Port:     proto.Uint32(uint32(namespacePort)),
					Path:     &healthCheck.Path,
					Statuses: []uint32{uint32(200), uint32(201), uint32(301), uint32(302)},
				},
			}
		}

		if protocol == "tcp" {
			builder.taskInfo.HealthCheck = &mesos.HealthCheck{
				Type: mesos.HealthCheck_TCP.Enum(),
				Tcp: &mesos.HealthCheck_TCPCheckInfo{
					Port: proto.Uint32(uint32(namespacePort)),
				},
			}
		}
	}

	builder.taskInfo.HealthCheck.IntervalSeconds = proto.Float64(healthCheck.IntervalSeconds)
	builder.taskInfo.HealthCheck.TimeoutSeconds = proto.Float64(healthCheck.TimeoutSeconds)
	builder.taskInfo.HealthCheck.ConsecutiveFailures = proto.Uint32(healthCheck.ConsecutiveFailures)
	builder.taskInfo.HealthCheck.GracePeriodSeconds = proto.Float64(healthCheck.GracePeriodSeconds)
	builder.taskInfo.HealthCheck.DelaySeconds = proto.Float64(healthCheck.DelaySeconds)

	return builder
}
