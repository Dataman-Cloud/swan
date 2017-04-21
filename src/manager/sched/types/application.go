package types

import "time"

type Application struct {
	ID string
	Meta
	Spec           ApplicationSpec
	PreviousSpecId string
	Endpoint       Endpoint
	UpdateStatus   UpdateStatus
}

type ApplicationSpec struct {
	Annotations
	Ip        []string
	Mode      string
	Instances int32
	RunAs     string

	TaskTemplate TaskSpec
	UpdatePolicy *UpdatePolicy
	Constraints  [][]string
}

type Version struct {
	ID    string
	AppId string
	Meta

	Spec ApplicationSpec
}

// Endpoint use to provide app endpoint info by service discovery
type Endpoint struct {
}

const (
	// UpdateStateUpdating is the updating state.
	UpdateStateUpdating UpdateState = "updating"
	// UpdateStatePaused is the paused state.
	UpdateStatePaused UpdateState = "paused"
	// UpdateStateCompleted is the completed state.
	UpdateStateCompleted UpdateState = "completed"
)

// UpdateStatus use to record application middle status
type UpdateStatus struct {
	State       UpdateState `json:",omitempty"`
	StartedAt   time.Time   `json:",omitempty"`
	CompletedAt time.Time   `json:",omitempty"`
	Message     string      `json:",omitempty"`
}

type Task struct {
	ID string
	Annotations
	AppId         string
	OfferId       string
	AgentId       string
	AgentHostname string
	Status        string

	Spec TaskSpec
}

type TaskSpec struct {
	Command      string
	Cpus         float64
	Disk         float64
	Mem          float64
	Env          map[string]string
	HealthChecks []*HealthCheck
	KillPolicy   *KillPolicy
	Uris         []string
	Container    *Container
}

type Container struct {
	Type    string
	Docker  *Docker
	Volumes []*Volume
}

type Docker struct {
	ForcePullImage bool
	Image          string
	Network        string
	Parameters     []*Parameter
	PortMappings   []*PortMapping
	Privileged     bool
}

type Parameter struct {
	Key   string
	Value string
}

type PortMapping struct {
	ContainerPort int32
	Name          string
	Protocol      string
}

type Volume struct {
	ContainerPath string
	HostPath      string
	Mode          string
}

type KillPolicy struct {
	Duration int64
}
