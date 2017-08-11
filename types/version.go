package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/Dataman-Cloud/swan/utils"
)

const (
	// update onfailure action
	UpdateStop     = "stop"
	UpdateContinue = "continue"
	UpdateRollback = "rollback" // TODO(nmg)
)

type VersionList []*Version

func (vl VersionList) Len() int      { return len(vl) }
func (vl VersionList) Swap(i, j int) { vl[i], vl[j] = vl[j], vl[i] }
func (vl VersionList) Less(i, j int) bool {
	m, _ := strconv.Atoi(vl[i].ID)
	n, _ := strconv.Atoi(vl[j].ID)

	return m < n
}

func (vl VersionList) Sort() {
	sort.Sort(vl)
}

func (vl VersionList) Reverse() {
	sort.Sort(sort.Reverse(vl))
}

type Version struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Command       string            `json:"cmd"`
	CPUs          float64           `json:"cpus"`
	GPUs          float64           `json:"gpus"`
	Mem           float64           `json:"mem"`
	Disk          float64           `json:"disk"`
	Instances     int32             `json:"instances"`
	RunAs         string            `json:"runAs"`
	Cluster       string            `json:"cluster"`
	Container     *Container        `json:"container"`
	Labels        map[string]string `json:"labels"`
	HealthCheck   *HealthCheck      `json:"healthCheck"`
	Env           map[string]string `json:"env"`
	KillPolicy    *KillPolicy       `json:"kill"`
	RestartPolicy *RestartPolicy    `json:"restart"`
	UpdatePolicy  *UpdatePolicy     `json:"update"`
	Constraints   []*Constraint     `json:"constraints"`
	URIs          []string          `json:"uris"`
	IPs           []string          `json:"ips"`
	Proxy         *Proxy            `json:"proxy"`
}

type Container struct {
	Type    string    `json:"type"`
	Docker  *Docker   `json:"docker"`
	Volumes []*Volume `json:"volumes"`
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

func (pm *PortMapping) Valid() error {
	if pm.Name == "" {
		return errors.New("port must be named")
	}
	if pm.HostPort < 0 || pm.HostPort > 65535 {
		return errors.New("host port out of range")
	}
	if pm.ContainerPort < 0 || pm.ContainerPort > 65535 {
		return errors.New("container port out of range")
	}
	if p := strings.ToLower(pm.Protocol); p != "tcp" && p != "udp" {
		return errors.New("unsupported port protocol")
	}
	return nil
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
	Step      int64   `json:"step"` // TODO(nmg)
	Delay     float64 `json:"delay"`
	OnFailure string  `json:"onFailure,omitempty"`
}

type HealthCheck struct {
	Protocol            string  `json:"protocol,omitempty"`
	PortName            string  `json:"portName,omitempty"`
	Path                string  `json:"path,omitempty"`
	Command             string  `json:"cmd,omitempty"`
	ConsecutiveFailures uint32  `json:"consecutiveFailures,omitempty"`
	GracePeriodSeconds  float64 `json:"gracePeriodSeconds,omitempty"`
	IntervalSeconds     float64 `json:"intervalSeconds,omitempty"`
	TimeoutSeconds      float64 `json:"timeoutSeconds,omitempty"`
	DelaySeconds        float64 `json:"delaySeconds,omitempty"`
}

func (h *HealthCheck) Valid() error {
	if h == nil {
		return nil
	}

	if h.IsEmpty() {
		return nil
	}

	switch p := strings.ToLower(h.Protocol); p {
	case "cmd":
		if h.Command == "" {
			return errors.New("command required for cmd health check")
		}
	case "tcp":
		if h.PortName == "" {
			return errors.New("port name required for tcp health check")
		}

	case "http":
		if h.PortName == "" {
			return errors.New("port name required for http health check")
		}
		if h.Path == "" {
			return errors.New("path required for http health check")
		}

	default:
		return errors.New("unsupported health check protocol")
	}

	if h.ConsecutiveFailures < 0 {
		return errors.New("consecutiveFailures can't be negative")
	}

	if h.GracePeriodSeconds < 0 {
		return errors.New("gracePeriodSeconds can't be negative")
	}

	if h.IntervalSeconds < 0 {
		return errors.New("intervalSeconds can't be negative")
	}

	if h.TimeoutSeconds < 0 {
		return errors.New("timeoutSeconds can't be negative")
	}

	if h.DelaySeconds < 0 {
		return errors.New("delaySeconds can't be negative")
	}

	return nil
}

func (h *HealthCheck) IsEmpty() bool {
	return h.Protocol == "" &&
		h.PortName == "" &&
		h.Path == "" &&
		h.Command == "" &&
		h.ConsecutiveFailures == 0 &&
		h.GracePeriodSeconds == 0 &&
		h.IntervalSeconds == 0 &&
		h.TimeoutSeconds == 0 &&
		h.DelaySeconds == 0
}

type Proxy struct {
	Enabled bool   `json:"enabled"`
	Alias   string `json:"alias"`
	Listen  string `json:"listen"`
	Sticky  bool   `json:"sticky"`
}

// similiar as above, but `Listen` int type
type ProxyAlias struct {
	Enabled bool   `json:"enabled"`
	Alias   string `json:"alias"`
	Listen  int    `json:"listen"`
	Sticky  bool   `json:"sticky"`
}

// hijack Marshaler & Unmarshaler to make fit with int type `Listen`
func (p *Proxy) MarshalJSON() ([]byte, error) {
	var l int
	if p.Listen != "" {
		l, _ = strconv.Atoi(strings.TrimPrefix(p.Listen, ":"))
	}
	var pa = &ProxyAlias{
		Enabled: p.Enabled,
		Alias:   p.Alias,
		Listen:  l,
		Sticky:  p.Sticky,
	}
	return json.Marshal(pa)
}

func (p *Proxy) UnmarshalJSON(data []byte) error {
	var pa ProxyAlias
	if err := json.Unmarshal(data, &pa); err != nil {
		return err
	}
	p.Enabled = pa.Enabled
	p.Alias = pa.Alias
	if pa.Listen > 0 {
		p.Listen = fmt.Sprintf(":%d", pa.Listen)
	}
	p.Sticky = pa.Sticky
	return nil
}

func (v *Version) IsHealthSet() bool {
	return v.HealthCheck != nil && !v.HealthCheck.IsEmpty()
}

// AddLabel adds a label to the application
func (v *Version) AddLabel(name, value string) *Version {
	if v.Labels == nil {
		v.EmptyLabels()
	}
	v.Labels[name] = value

	return v
}

// EmptyLabels explicitly empties the labels
func (v *Version) EmptyLabels() *Version {
	v.Labels = map[string]string{}

	return v
}

func (v *Version) Validate() error {
	if v.Container == nil {
		return errors.New("swan only support mesos docker containerization, no container found")
	}

	if v.Container.Docker == nil {
		return errors.New("swan only support mesos docker containerization, no container found")
	}

	if v.Container.Docker.Image == "" {
		return errors.New("image field required")
	}

	if n := len(v.Name); n == 0 || n > 63 {
		return errors.New("invalid appName: appName empty or too long")
	}

	if len(v.RunAs) == 0 {
		return errors.New("runAs should not empty")
	}

	if v.Instances <= 0 {
		return errors.New("invalid instances: instances must be specified and should greater than 0")
	}

	if v.Mem < 0 {
		return errors.New("memory can't be negative")
	}

	if v.CPUs < 0 {
		return errors.New("cpus can't be negative")
	}

	if v.GPUs < 0 {
		return errors.New("gpus can't be negative")
	}

	if v.Disk < 0 {
		return errors.New("disk can't be negative")
	}

	for _, cons := range v.Constraints {
		if err := cons.validate(); err != nil {
			return err
		}
	}

	if err := utils.LegalDomain(v.Name); err != nil {
		return err
	}

	if err := utils.LegalDomain(v.RunAs); err != nil {
		return err
	}

	switch network := strings.ToLower(v.Container.Docker.Network); network {

	case "host", "bridge":
		// ensure port name uniq
		seen := make(map[string]bool)
		for _, pm := range v.Container.Docker.PortMappings {
			if err := pm.Valid(); err != nil {
				return err
			}
			if _, ok := seen[pm.Name]; ok {
				return fmt.Errorf("port name %s conflict", pm.Name)
			}
			seen[pm.Name] = true
		}

		// health check validation
		if err := v.HealthCheck.Valid(); err != nil {
			return err
		}

	default:
		if len(v.IPs) != int(v.Instances) {
			return fmt.Errorf("IPs(%d) should match instances(%d)", len(v.IPs), v.Instances)
		}

		// ensure ip valid & uniq
		seen := make(map[string]bool)
		for _, ip := range v.IPs {
			if addr := net.ParseIP(ip); addr == nil || addr.IsLoopback() {
				return errors.New("invalid ip: " + ip)
			}
			if _, ok := seen[ip]; ok {
				return fmt.Errorf("ip %s conflict", ip)
			}
			seen[ip] = true
		}

		if v.IsHealthSet() {
			if p := strings.ToLower(v.HealthCheck.Protocol); p == "cmd" {
				return fmt.Errorf("can't use cmd health check on fixed type app")
			}
		}
	}

	return nil
}
