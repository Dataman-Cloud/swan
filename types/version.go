package types

type Version struct {
	ID           string             `json:"id"`
	Command      *string            `json:"cmd"`
	Cpus         float64            `json:"cpus"`
	Mem          float64            `json:"mem"`
	Disk         float64            `json:"disk"`
	Instances    int                `json:"instances"`
	Container    *Container         `json:"container"`
	Labels       *map[string]string `json:"labels"`
	HealthChecks []*HealthCheck     `json:"healthChecks"`
	Env          map[string]string  `json:"env"`
	KillPolicy   *KillPolicy        `json:"killPolicy"`
	UpdatePolicy *UpdatePolicy      `json:"updatePolicy"`
}

// Container is the definition for a container type in marathon
type Container struct {
	Type    string    `json:"type,omitempty"`
	Docker  *Docker   `json:"docker,omitempty"`
	Volumes []*Volume `json:"volumes,omitempty"`
}

// Docker is the docker definition from a marathon application
type Docker struct {
	ForcePullImage *bool          `json:"forcePullImage,omitempty"`
	Image          *string        `json:"image,omitempty"`
	Network        string         `json:"network,omitempty"`
	Parameters     *[]Parameter   `json:"parameters,omitempty"`
	PortMappings   *[]PortMapping `json:"portMappings,omitempty"`
	Privileged     *bool          `json:"privileged,omitempty"`
}

type Parameter struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type PortMapping struct {
	ContainerPort int    `json:"containerPort,omitempty"`
	Name          string `json:"name,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
}

type Volume struct {
	ContainerPath string `json:"containerPath,omitempty"`
	HostPath      string `json:"hostPath,omitempty"`
	Mode          string `json:"mode,omitempty"`
}

type KillPolicy struct {
	Duration int64 `json:"duration"`
}
