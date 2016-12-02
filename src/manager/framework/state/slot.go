package state

import (
	"fmt"
	"sync"

	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
)

const (
	SLOT_STATE_PENDING_OFFER     = "pending_offer"
	SLOT_STATE_TASK_DISPATCHED   = "task_dispatched"
	SLOT_STATE_TASK_RUNING       = "task_running"
	SLOT_STATE_TASK_FAILED       = "task_need_reschedue"
	SLOT_STATE_TASK_RESCHEDULING = "task_rescheduling"
	SLOT_STATE_TASK_STOPPING     = "task_stopping"
	SLOT_STATE_TASK_STOPPED      = "task_stopped"
)

type Slot struct {
	Id      int
	Name    string
	App     *App
	Version *types.Version
	State   string

	RunningTask *Task
	Tasks       []*Task

	resourceReservationLock sync.Mutex
}

func NewSlot(app *App, version *types.Version, index int) *Slot {
	slot := &Slot{
		Id:      index,
		App:     app,
		Version: version,
		Tasks:   make([]*Task, 0),
		State:   SLOT_STATE_PENDING_OFFER,
		Name:    fmt.Sprintf("%d-%s", index, version.AppId), // should be app.AppId

		resourceReservationLock: sync.Mutex{},
	}

	return slot
}

func (slot *Slot) TestOfferMatch(ow *OfferWrapper) bool {
	return ow.CpuRemain() > slot.Version.Cpus &&
		ow.MemRemain() > slot.Version.Mem &&
		ow.DiskRemain() > slot.Version.Disk
}

func (slot *Slot) ReserveOfferAndPrepareTaskInfo(ow *OfferWrapper) (*OfferWrapper, *mesos.TaskInfo) {
	slot.resourceReservationLock.Lock()
	defer slot.resourceReservationLock.Unlock()

	ow.CpusUsed += slot.Version.Cpus
	ow.MemUsed += slot.Version.Mem
	ow.DiskUsed += slot.Version.Disk

	return ow, slot.BuildTaskInfo(ow.Offer)
}

func (slot *Slot) Resources() []*mesos.Resource {
	var resources = []*mesos.Resource{}

	if slot.Version.Cpus > 0 {
		resources = append(resources, createScalarResource("cpus", slot.Version.Cpus))
	}

	if slot.Version.Mem > 0 {
		resources = append(resources, createScalarResource("mem", slot.Version.Cpus))
	}

	if slot.Version.Disk > 0 {
		resources = append(resources, createScalarResource("disk", slot.Version.Disk))
	}

	return resources
}

func (slot *Slot) SetState(state string) {
	slot.State = state

	switch slot.State {
	case SLOT_STATE_TASK_DISPATCHED:
		fmt.Println(slot.State)
	case SLOT_STATE_TASK_RUNING:
		fmt.Println(slot.State)
	case SLOT_STATE_TASK_FAILED:
		fmt.Println(slot.State)
		// restart if needed
	default:
	}
	// persist to db
}

func (slot *Slot) BuildTaskInfo(offer *mesos.Offer) *mesos.TaskInfo {
	logrus.Infof("Prepared task for launch with offer %s", *offer.GetId().Value)
	taskInfo := mesos.TaskInfo{
		Name: proto.String(slot.Name),
		TaskId: &mesos.TaskID{
			Value: proto.String(slot.Name),
		},
		AgentId:   offer.AgentId,
		Resources: slot.Resources(),
		Command: &mesos.CommandInfo{
			Shell: proto.Bool(false),
			Value: nil,
		},
		Container: &mesos.ContainerInfo{
			Type: mesos.ContainerInfo_DOCKER.Enum(),
			Docker: &mesos.ContainerInfo_DockerInfo{
				Image: &slot.Version.Container.Docker.Image,
			},
		},
	}

	taskInfo.Container.Docker.Privileged = &slot.Version.Container.Docker.Privileged
	taskInfo.Container.Docker.ForcePullImage = &slot.Version.Container.Docker.ForcePullImage

	for _, parameter := range slot.Version.Container.Docker.Parameters {
		taskInfo.Container.Docker.Parameters = append(taskInfo.Container.Docker.Parameters, &mesos.Parameter{
			Key:   proto.String(parameter.Key),
			Value: proto.String(parameter.Value),
		})
	}

	for _, volume := range slot.Version.Container.Volumes {
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
	for k, v := range slot.Version.Env {
		vars = append(vars, &mesos.Environment_Variable{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}

	taskInfo.Command.Environment = &mesos.Environment{
		Variables: vars,
	}

	if slot.Version.Labels != nil {
		labels := make([]*mesos.Label, 0)
		for k, v := range slot.Version.Labels {
			labels = append(labels, &mesos.Label{
				Key:   proto.String(k),
				Value: proto.String(v),
			})
		}

		taskInfo.Labels = &mesos.Labels{
			Labels: labels,
		}
	}

	switch slot.Version.Container.Docker.Network {
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
		//for _, m := range task.PortMappings {
		//hostPort := ports[s.TaskLaunched]
		//taskInfo.Container.Docker.PortMappings = append(taskInfo.Container.Docker.PortMappings,
		//&mesos.ContainerInfo_DockerInfo_PortMapping{
		//HostPort:      proto.Uint32(uint32(hostPort)),
		//ContainerPort: proto.Uint32(m.Port),
		//Protocol:      proto.String(m.Protocol),
		//},
		//)
		//taskInfo.Resources = append(taskInfo.Resources, &mesos.Resource{
		//Name: proto.String("ports"),
		//Type: mesos.Value_RANGES.Enum(),
		//Ranges: &mesos.Value_Ranges{
		//Range: []*mesos.Value_Range{
		//{
		//Begin: proto.Uint64(uint64(hostPort)),
		//End:   proto.Uint64(uint64(hostPort)),
		//},
		//},
		//},
		//})
		//}
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_BRIDGE.Enum()
	default:
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_NONE.Enum()
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
