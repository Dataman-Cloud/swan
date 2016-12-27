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
	Id         string
	TaskInfoId string
	Version    *types.Version
	Slot       *Slot

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

	Created     time.Time
	taskBuilder *TaskBuilder
}

func NewTask(version *types.Version, slot *Slot) *Task {
	task := &Task{
		Id:        strings.Replace(uuid.NewV4().String(), "-", "", -1),
		Version:   version,
		Slot:      slot,
		HostPorts: make([]uint64, 0),
		Created:   time.Now(),
	}

	task.TaskInfoId = fmt.Sprintf("%s-%s", task.Slot.Id, task.Id)

	return task
}

func (task *Task) PrepareTaskInfo(ow *OfferWrapper) *mesos.TaskInfo {
	offer := ow.Offer
	logrus.Infof("Prepared task %s for launch with offer %s", task.Slot.Id, *offer.GetId().Value)

	versionSpec := task.Slot.Version
	containerSpec := task.Slot.Version.Container
	dockerSpec := task.Slot.Version.Container.Docker

	task.taskBuilder = NewTaskBuilder(task)
	task.taskBuilder.SetName(task.TaskInfoId).SetTaskId(task.TaskInfoId).SetAgentId(*offer.GetAgentId().Value)
	task.taskBuilder.SetResources(task.Slot.ResourcesNeeded())
	task.taskBuilder.SetCommand(false, "")

	task.taskBuilder.SetContainerType("docker").SetContainerDockerImage(dockerSpec.Image).
		SetContainerDockerPrivileged(dockerSpec.Privileged).
		SetContainerDockerForcePullImage(dockerSpec.ForcePullImage).
		AppendContainerDockerVolumes(containerSpec.Volumes)

	task.taskBuilder.AppendContainerDockerEnvironments(versionSpec.Env).SetURIs(versionSpec.Uris).SetLabels(versionSpec.Labels)

	task.taskBuilder.AppendContainerDockerParameters(task.Slot.Version.Container.Docker.Parameters)

	if task.Slot.App.IsFixed() {
		ipParameter := types.Parameter{
			Key:   "ip",
			Value: task.Slot.Ip,
		}
		task.taskBuilder.AppendContainerDockerParameters([]*types.Parameter{&ipParameter})
	}

	task.taskBuilder.SetNetwork(dockerSpec.Network, ow.PortsRemain())
	task.HostPorts = task.taskBuilder.HostPorts

	return task.taskBuilder.taskInfo
}

func (task *Task) Kill() {
	logrus.Infof("Kill task %s", task.Slot.Id)
	call := &sched.Call{
		FrameworkId: mesos_connector.Instance().Framework.GetId(),
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

	mesos_connector.Instance().MesosCallChan <- call
}
