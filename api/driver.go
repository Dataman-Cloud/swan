package api

import (
	"io"

	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/mole"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/andygrunwald/megos"
)

type Driver interface {
	KillTask(taskId string, agentId string, gradePeriod int64) error
	LaunchTasks([]*mesos.Task) error

	// kvm stop/start/suspend/resume
	StopKvmTask(taskId, agentId, executorId string) error
	StartKvmTask(taskId, agentId, executorId string) error
	// SuspendKvmTask(taskId string, agentId string) error
	// ResumeKvmTask(taskId string, agentId string) error

	ClusterName() string

	SubscribeEvent(io.Writer, string) error
	FullTaskEventsAndRecords() []*types.CombinedEvents
	SendEvent(string, *types.Task) error

	ClusterAgents() map[string]*mole.ClusterAgent
	ClusterAgent(id string) *mole.ClusterAgent
	CloseClusterAgent(id string)

	MesosState() (*megos.State, error)

	// for debug convenience
	Dump() interface{}
	Load() map[string]interface{}
	FrameworkInfo() *types.FrameworkInfo
}
