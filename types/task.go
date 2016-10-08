package types

import (
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
)

var DefaultTaskState = mesos.TaskState_TASK_STAGING

type Task struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Command      string          `json:"cmd"`
	Cpus         float64         `json:"cpus,string"`
	Disk         float64         `json:"disk,string"`
	Mem          float64         `json:"mem,string"`
	Image        string          `json:"image"`
	Network      string          `json:"network"`
	PortMappings []*PortMappings `json:"port_mappings"`
	HealthCheck  *HealthCheck    `json:"health_check"`

	AgentId       string           `json:"agent_id,string"`
	AgentHostname string           `json:"agent_hostname"`
	State         *mesos.TaskState `json:"state"`
}

type PortMappings struct {
	Port     uint32 `json:"port"`
	Protocol string `json:"protocol"`
}

type HealthCheck struct {
	Protocol string `json:"protocol"`
	Port     uint32 `json:"port"`
}
