package types

type Version struct {
	ID           string            `json:"id,omitempty"`
	AppName      string            `json:"appName,omitempty"`
	AppVersion   string            `json:"appVersion,omitempty"`
	Command      string            `json:"cmd,omitempty"`
	Args         []string          `json:"args,omitempty"`
	CPUs         float64           `json:"cpus"`
	Mem          float64           `json:"mem"`
	Disk         float64           `json:"disk"`
	Instances    int32             `json:"instances"`
	RunAs        string            `json:"runAs"`
	Priority     int32             `json:"priority"`
	Container    *Container        `json:"container"`
	Labels       map[string]string `json:"labels,omitempty"`
	HealthCheck  *HealthCheck      `json:"healthCheck,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	KillPolicy   *KillPolicy       `json:"killPolicy,omitempty"`
	UpdatePolicy *UpdatePolicy     `json:"updatPolicy,omitempty"`
	Constraints  string            `json:"constraints,omitempty"`
	URIs         []string          `json:"uris,omitempty"`
	IP           []string          `json:"ip,omitempty"`
}

type Container struct {
	Type    string    `json:"type"`
	Docker  *Docker   `json:"docker"`
	Volumes []*Volume `json:"volumes,omitempty"`
}

type Docker struct {
	ForcePullImage bool           `json:"forcePullImage,omitempty"`
	Image          string         `json:"image"`
	Network        string         `json:"network,omitempty"`
	Parameters     []*Parameter   `json:"parameters,omitempty"`
	PortMappings   []*PortMapping `json:"portMappings,omitempty"`
	Privileged     bool           `json:"privileged,omitempty"`
}

type Parameter struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type PortMapping struct {
	ContainerPort int32  `json:"containerPort,omitempty"`
	HostPort      int32  `json:"hostPort,omitempty"`
	Name          string `json:"name,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
}

type Volume struct {
	ContainerPath string `json:"containerPath,omitempty"`
	HostPath      string `json:"hostPath,omitempty"`
	Mode          string `json:"mode,omitempty"`
}

type KillPolicy struct {
	Duration int64 `json:"duration,omitempty"`
}

type UpdatePolicy struct {
	UpdateDelay  int32  `json:"updateDelay,omitempty"`
	MaxRetries   int32  `json:"maxRetries,omitempty"`
	MaxFailovers int32  `json:"maxFailovers,omitempty"`
	Action       string `json:"action,omitempty"`
}

type HealthCheck struct {
	ID                  string  `json:"id,omitempty"`
	Address             string  `json:"address,omitempty"`
	TaskID              string  `json:"taskID,omitempty"`
	AppID               string  `json:"appID,omitempty"`
	Protocol            string  `json:"protocol,omitempty"`
	PortName            string  `json:"portName,omitempty"`
	Value               string  `json:"value,omitempty"`
	Path                string  `json:"path,omitempty"`
	ConsecutiveFailures uint32  `json:"consecutiveFailures,omitempty"`
	GracePeriodSeconds  float64 `json:"gracePeriodSeconds,omitempty"`
	IntervalSeconds     float64 `json:"intervalSeconds,omitempty"`
	TimeoutSeconds      float64 `json:"timeoutSeconds,omitempty"`
	DelaySeconds        float64 `json:"delaySeconds,omitempty"`
}

// AddLabel adds a label to the application
//		name:	the name of the label
//		value: value for this label
func (v *Version) AddLabel(name, value string) *Version {
	if v.Labels == nil {
		v.EmptyLabels()
	}
	v.Labels[name] = value

	return v
}

// EmptyLabels explicitly empties the labels -- use this if you need to empty
// the labels of an application that already has labels set (setting labels to nil will
// keep the current value)
func (v *Version) EmptyLabels() *Version {
	v.Labels = map[string]string{}

	return v
}
