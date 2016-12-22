package state

import (
	"fmt"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/framework/mesos_connector"
	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	uuid "github.com/satori/go.uuid"
)

const (
	SWAN_RESERVED_NETWORK = "swan"
)

type Task struct {
	Id             string
	TaskInfoId     string
	MesosConnector *mesos_connector.MesosConnector
	Version        *types.Version
	Slot           *Slot

	State  string
	Stdout string
	Stderr string

	HostPorts     []uint64
	OfferId       string
	AgentId       string
	Ip            string
	AgentHostName string

	Reason  string
	Message string
	Source  string

	Created time.Time
}

func NewTask(mesosConnector *mesos_connector.MesosConnector, version *types.Version, slot *Slot) *Task {
	task := &Task{
		Id:             strings.Replace(uuid.NewV4().String(), "-", "", -1),
		MesosConnector: mesosConnector,
		Version:        version,
		Slot:           slot,
		HostPorts:      make([]uint64, 0),
		Created:        time.Now(),
	}

	task.TaskInfoId = fmt.Sprintf("%s-%s", task.Slot.Id, task.Id)

	return task
}

func (task *Task) PrepareTaskInfo(ow *OfferWrapper) *mesos.TaskInfo {
	offer := ow.Offer

	logrus.Infof("Prepared task %s for launch with offer %s", task.Slot.Id, *offer.GetId().Value)

	taskInfo := mesos.TaskInfo{
		Name: proto.String(task.TaskInfoId),
		TaskId: &mesos.TaskID{
			Value: proto.String(task.TaskInfoId),
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

	// check if app run in fixed mode and has reserved enough IP
	if task.Slot.App.IsFixed() {
		taskInfo.Container.Docker.Parameters = append(taskInfo.Container.Docker.Parameters, &mesos.Parameter{
			Key:   proto.String("ip"),
			Value: proto.String(task.Slot.Ip),
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
		ports := ow.PortsRemain()
		if len(ports) < len(task.Slot.Version.Container.Docker.PortMappings) {
			logrus.Errorf("No ports resource defined")
			break
		}

		task.HostPorts = make([]uint64, 0)
		for index, m := range task.Slot.Version.Container.Docker.PortMappings {
			hostPort := ports[index]
			task.HostPorts = append(task.HostPorts, hostPort)
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

	case SWAN_RESERVED_NETWORK:
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_USER.Enum()
		taskInfo.Container.NetworkInfos = append(taskInfo.Container.NetworkInfos, &mesos.NetworkInfo{
			Name: proto.String(SWAN_RESERVED_NETWORK),
		})

	default:
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_NONE.Enum()
	}

	// setup task health check
	if task.Slot.App.IsReplicates() {
		if len(task.Slot.Version.HealthChecks) > 0 {
			for _, healthCheck := range task.Slot.Version.HealthChecks {
				var hostPort *uint32
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

				protocol := strings.ToLower(healthCheck.Protocol)
				if protocol == "http" {
					taskInfo.HealthCheck = &mesos.HealthCheck{
						Type: mesos.HealthCheck_HTTP.Enum(),
						Http: &mesos.HealthCheck_HTTPCheckInfo{
							Scheme:   proto.String("http"),
							Port:     hostPort,
							Path:     &healthCheck.Path,
							Statuses: []uint32{uint32(200), uint32(201), uint32(301), uint32(302)},
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
				taskInfo.HealthCheck.ConsecutiveFailures = proto.Uint32(healthCheck.ConsecutiveFailures)
				taskInfo.HealthCheck.GracePeriodSeconds = proto.Float64(healthCheck.GracePeriodSeconds)
			}
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
		FrameworkId: task.MesosConnector.Framework.GetId(),
		Type:        sched.Call_KILL.Enum(),
		Kill: &sched.Call_Kill{
			TaskId: &mesos.TaskID{
				Value: proto.String(task.TaskInfoId),
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

	task.MesosConnector.MesosCallChan <- call
}
