package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"path"
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

func (c *Container) Valid() error {
	if strings.ToLower(c.Type) != "docker" {
		return errors.New("only support docker containerization")
	}
	if c.Docker == nil {
		return errors.New("docker containerization settings required")
	}
	for _, v := range c.Volumes {
		if err := v.Valid(); err != nil {
			return err
		}
	}
	return c.Docker.Valid()
}

type Docker struct {
	ForcePullImage bool           `json:"forcePullImage,omitempty"`
	Image          string         `json:"image"`
	Network        string         `json:"network,omitempty"`
	Parameters     []*Parameter   `json:"parameters,omitempty"`
	PortMappings   []*PortMapping `json:"portMappings,omitempty"`
	Privileged     bool           `json:"privileged,omitempty"`
}

func (d *Docker) Valid() error {
	if d.Network == "" {
		return errors.New("network required")
	}
	if d.Image == "" {
		return errors.New("image required")
	}
	seen := make(map[string]bool)
	for _, pm := range d.PortMappings {
		if err := pm.Valid(); err != nil {
			return err
		}
		if _, ok := seen[pm.Name]; ok {
			return fmt.Errorf("port name %s conflict", pm.Name)
		}
		seen[pm.Name] = true
	}
	for _, p := range d.Parameters {
		if err := p.Valid(); err != nil {
			return err
		}
	}
	return nil
}

type Parameter struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func (p *Parameter) Valid() error {
	if p.Key == "" {
		return errors.New("Parameter.Key required")
	}
	return nil
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

func (v *Volume) Valid() error {
	if !path.IsAbs(v.HostPath) {
		return errors.New("Volume.HostPath should be absolute path")
	}
	switch v.Mode {
	case "RO", "RW":
	default:
		return errors.New("unsupported Volume.Mode, should be [RO,RW]")
	}
	return nil
}

type KillPolicy struct {
	Duration int64 `json:"duration,omitempty"` // by seconds
}

func (p *KillPolicy) Valid() error {
	if p.Duration < 0 {
		return errors.New("KillPolicy.Duration can't be negative")
	}
	return nil
}

type RestartPolicy struct {
	Retries int `json:"retries"`
}

func (p *RestartPolicy) Valid() error {
	if p.Retries < 0 {
		return errors.New("RestartPolicy.Retries can't be negative")
	}
	return nil
}

type UpdatePolicy struct {
	Delay     float64 `json:"delay"`
	OnFailure string  `json:"onFailure,omitempty"`
}

func (p *UpdatePolicy) Valid() error {
	if p.Delay < 0 {
		return errors.New("UpdatePolicy.Delay can't be negative")
	}
	switch p.OnFailure {
	case UpdateStop, UpdateContinue, UpdateRollback:
	default:
		return errors.New("unsupported OnFailure policy for update")
	}
	return nil
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

	if h.GracePeriodSeconds < 0 {
		return errors.New("gracePeriodSeconds can't be negative")
	}

	if h.IntervalSeconds <= 0 {
		return errors.New("intervalSeconds should be positive")
	}

	if h.TimeoutSeconds < 0 {
		return errors.New("timeoutSeconds can't be negative")
	}

	if h.DelaySeconds < 0 {
		return errors.New("delaySeconds can't be negative")
	}

	return nil
}

type Proxy struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Alias   string `json:"alias" yaml:"alias"`
	Listen  string `json:"listen" yaml:"listen"`
	Sticky  bool   `json:"sticky" yaml:"sticky"`
}

// similiar as above, but `Listen` int type
type ProxyAlias struct {
	Enabled bool   `json:"enabled"`
	Alias   string `json:"alias"`
	Listen  int    `json:"listen"`
	Sticky  bool   `json:"sticky"`
}

func (p *Proxy) Valid() error {
	if !p.Enabled {
		return nil
	}

	if p.Listen != "" {
		l, err := strconv.Atoi(strings.TrimPrefix(p.Listen, ":"))
		if err != nil {
			return err
		}
		if l < 0 || l > 65535 {
			return errors.New("proxy.Listen out of range")
		}
	}

	return nil
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
	switch {
	case pa.Listen == 0:
		p.Listen = ""
	case pa.Listen < 0 || pa.Listen > 65535:
		return errors.New("proxy.Listen out of range")
	default: // (0, 65535]
		p.Listen = fmt.Sprintf(":%d", pa.Listen)
	}
	p.Sticky = pa.Sticky
	return nil
}

func (v *Version) IsHealthSet() bool {
	return v.HealthCheck != nil
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
	// verify name
	if n := len(v.Name); n == 0 || n > 64 {
		return errors.New("appName should between (0,64]")
	}

	if err := utils.LegalDomain(v.Name); err != nil {
		return err
	}

	// verify resources
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

	if v.Instances <= 0 {
		return errors.New("instance count must be positive")
	}

	// verify runas
	if n := len(v.RunAs); n == 0 || n > 64 {
		return errors.New("runAs length should between (0,64]")
	}

	if err := utils.LegalDomain(v.RunAs); err != nil {
		return err
	}

	// verify cluster
	if n := len(v.Cluster); n > 64 {
		return errors.New("cluster name length should between [0,64]")
	}

	if err := utils.LegalDomain(v.Cluster); err != nil {
		return err
	}

	// verify container
	if v.Container == nil {
		return errors.New("container settings required")
	}

	if err := v.Container.Valid(); err != nil {
		return err
	}

	// verify healthcheck
	if v.IsHealthSet() {
		if err := v.HealthCheck.Valid(); err != nil {
			return err
		}
	}

	// verify killpolicy
	if v.KillPolicy != nil {
		if err := v.KillPolicy.Valid(); err != nil {
			return err
		}
	}

	// verify restartpolicy
	if v.RestartPolicy != nil {
		if err := v.RestartPolicy.Valid(); err != nil {
			return err
		}
	}

	// verify updatepolicy
	if v.UpdatePolicy != nil {
		if err := v.UpdatePolicy.Valid(); err != nil {
			return err
		}
	}

	// verify constraints
	for _, cons := range v.Constraints {
		if err := cons.validate(); err != nil {
			return err
		}
	}

	// verify proxy
	if v.Proxy != nil {
		if err := v.Proxy.Valid(); err != nil {
			return err
		}
	}

	// verify static ip mode
	switch network := strings.ToLower(v.Container.Docker.Network); network {
	case "host", "bridge":
	default:
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

		// ensure not cmd healthcheck
		if v.IsHealthSet() {
			if p := strings.ToLower(v.HealthCheck.Protocol); p == "cmd" {
				return fmt.Errorf("can't use cmd health check on fixed type app")
			}
		}
	}

	return nil
}
