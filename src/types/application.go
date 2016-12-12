package types

type Version struct {
	ID           string
	AppId        string
	Command      string
	Cpus         float64
	Mem          float64
	Disk         float64
	Instances    int32
	RunAs        string
	Container    *Container
	Labels       map[string]string
	HealthChecks []*HealthCheck
	Env          map[string]string
	KillPolicy   *KillPolicy
	UpdatePolicy *UpdatePolicy
	Constraints  []string
	Uris         []string
	Ip           []string
	Mode         string
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

type UpdatePolicy struct {
	UpdateDelay  int32
	MaxRetries   int32
	MaxFailovers int32
	Action       string
}

type HealthCheck struct {
	ID                     string
	Address                string
	TaskID                 string
	AppID                  string
	Protocol               string
	Port                   int32
	PortIndex              int32
	PortName               string
	Command                *Command
	Path                   string
	MaxConsecutiveFailures uint32
	GracePeriodSeconds     float64
	IntervalSeconds        float64
	TimeoutSeconds         float64
}

type Command struct {
	Value string
}
