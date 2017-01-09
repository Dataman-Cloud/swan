package state

import (
	"fmt"
	"strconv"
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
	ID         string
	TaskInfoID string
	Version    *types.Version
	Slot       *Slot

	State  string
	Stdout string
	Stderr string

	HostPorts     []uint64
	OfferID       string
	AgentID       string
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
		ID:        strings.Replace(uuid.NewV4().String(), "-", "", -1),
		Version:   version,
		Slot:      slot,
		HostPorts: make([]uint64, 0),
		Created:   time.Now(),
	}

	task.TaskInfoID = fmt.Sprintf("%s-%s", task.Slot.ID, task.ID)

	return task
}

func (task *Task) PrepareTaskInfo(ow *OfferWrapper) *mesos.TaskInfo {
	defaultLabels := make(map[string]string)
	defaultLabels["USER_ID"] = task.Slot.Version.RunAs
	defaultLabels["CLUSTER_ID"] = task.Slot.App.ClusterID
	defaultLabels["SLOT_ID"] = strconv.Itoa(task.Slot.Index)
	defaultLabels["APP_ID"] = task.Slot.App.ID
	defaultLabels["TASK_ID"] = task.TaskInfoID

	offer := ow.Offer
	logrus.Infof("Prepared task %s for launch with offer %s", task.Slot.ID, *offer.GetId().Value)

	versionSpec := task.Slot.Version
	containerSpec := task.Slot.Version.Container
	dockerSpec := task.Slot.Version.Container.Docker

	task.taskBuilder = NewTaskBuilder(task)
	task.taskBuilder.SetName(task.TaskInfoID).SetTaskId(task.TaskInfoID).SetAgentId(*offer.GetAgentId().Value)
	task.taskBuilder.SetResources(task.Slot.ResourcesNeeded())
	task.taskBuilder.SetCommand(false, task.Slot.Version.Command, task.Slot.Version.Args)

	task.taskBuilder.SetContainerType("docker").SetContainerDockerImage(dockerSpec.Image).
		SetContainerDockerPrivileged(dockerSpec.Privileged).
		SetContainerDockerForcePullImage(dockerSpec.ForcePullImage).
		AppendContainerDockerVolumes(containerSpec.Volumes)

	task.taskBuilder.AppendContainerDockerEnvironments(versionSpec.Env).SetURIs(versionSpec.URIs).AppendTaskInfoLabels(versionSpec.Labels)
	task.taskBuilder.AppendTaskInfoLabels(defaultLabels)

	task.taskBuilder.AppendContainerDockerParameters(task.Slot.Version.Container.Docker.Parameters)

	if task.Slot.App.IsFixed() {
		ipParameter := types.Parameter{
			Key:   "ip",
			Value: task.Slot.Ip,
		}
		task.taskBuilder.AppendContainerDockerParameters([]*types.Parameter{&ipParameter})
	}

	for k, v := range defaultLabels {
		p := types.Parameter{
			Key:   "label",
			Value: fmt.Sprintf("%s=%s", k, v),
		}
		task.taskBuilder.AppendContainerDockerParameters([]*types.Parameter{&p})
	}

	task.taskBuilder.SetNetwork(dockerSpec.Network, ow.PortsRemain()).SetHealthCheck(versionSpec.HealthChecks)
	task.HostPorts = task.taskBuilder.HostPorts

	return task.taskBuilder.taskInfo
}

func (task *Task) Kill() {
	logrus.Infof("Kill task %s", task.Slot.ID)
	call := &sched.Call{
		FrameworkId: mesos_connector.Instance().Framework.GetId(),
		Type:        sched.Call_KILL.Enum(),
		Kill: &sched.Call_Kill{
			TaskId: &mesos.TaskID{
				Value: proto.String(task.TaskInfoID),
			},
			AgentId: &mesos.AgentID{
				Value: &task.AgentID,
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
