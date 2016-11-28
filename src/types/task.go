package types

import (
	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
)

var DefaultTaskState = mesos.TaskState_TASK_STAGING

type Task struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	Command        *string            `json:"cmd"`
	Cpus           float64            `json:"cpus"`
	Disk           float64            `json:"disk"`
	Mem            float64            `json:"mem"`
	Image          *string            `json:"image"`
	Network        string             `json:"network"`
	PortMappings   []*PortMappings    `json:"port_mappings"`
	Privileged     *bool              `json:"privileged"`
	Parameters     []*Parameter       `json:"parameters"`
	ForcePullImage *bool              `json:"forcePullImage"`
	Volumes        []*Volume          `json:"volumes"`
	Env            map[string]string  `json:"env"`
	Labels         *map[string]string `json:"labels"`
	HealthChecks   []*HealthCheck     `json:"healthChecks"`

	OfferId       *string `json:"offer_id"`
	AgentId       *string `json:"agent_id,string"`
	AgentHostname *string `json:"agent_hostname"`
	Status        string  `json:"status"`
	AppId         string  `json:"app_id"`

	KillPolicy *KillPolicy `json:"kill_policy"`
}

type PortMappings struct {
	Port     uint32 `json:"port"`
	Protocol string `json:"protocol"`
}
