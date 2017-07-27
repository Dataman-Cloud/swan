package types

import (
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
	Mem           float64           `json:"mem"`
	Disk          float64           `json:"disk"`
	Instances     int32             `json:"instances"`
	RunAs         string            `json:"runAs"`
	Priority      int32             `json:"priority"`
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
	Command             string  `json:"cmd, omitempty"`
	ConsecutiveFailures uint32  `json:"consecutiveFailures,omitempty"`
	GracePeriodSeconds  float64 `json:"gracePeriodSeconds,omitempty"`
	IntervalSeconds     float64 `json:"intervalSeconds,omitempty"`
	TimeoutSeconds      float64 `json:"timeoutSeconds,omitempty"`
	DelaySeconds        float64 `json:"delaySeconds,omitempty"`
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

type Gateway struct {
	Enabled bool    `json:"enabled"`
	Weight  float64 `json:"weight,omitempty"`
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

// validate version
// TODO(nmg): use json schema validation replace latter.
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

	if v.CPUs < 0.01 {
		return errors.New("cpu should >= 0.01")
	}
	if v.Mem < 5 {
		return errors.New("mem should >= 5m")
	}

	// FIXME(nmg)
	if len(v.Constraints) != 0 {
		for _, cons := range v.Constraints {
			if err := cons.validate(); err != nil {
				return err
			}
		}
	}

	if err := utils.LegalDomain(v.Name); err != nil {
		return err
	}

	if err := utils.LegalDomain(v.RunAs); err != nil {
		return err
	}

	network := strings.ToLower(v.Container.Docker.Network)

	if network != "host" && network != "bridge" {
		if len(v.IPs) != int(v.Instances) {
			return fmt.Errorf("Ip number must equal instance number. required: %d actual: %d", v.Instances, len(v.IPs))
		}

		for _, ip := range v.IPs {
			if addr := net.ParseIP(ip); addr == nil || addr.IsLoopback() {
				return errors.New("invalid ip: " + ip)
			}
		}

		if v.HealthCheck != nil {
			protocol := v.HealthCheck.Protocol
			v := strings.ToLower(utils.StripSpaces(protocol))
			if v == "cmd" {
				return fmt.Errorf("doesn't supported protocol %s for health check for fixed type app", protocol)
			}
		}
	} else {
		for _, portmapping := range v.Container.Docker.PortMappings {
			if strings.TrimSpace(portmapping.Name) == "" {
				return errors.New("each port mapping should have a unique identified name")
			}
		}

		portNames := make([]string, 0)
		for _, portmapping := range v.Container.Docker.PortMappings {
			portNames = append(portNames, portmapping.Name)
		}

		if !utils.SliceUnique(portNames) {
			return errors.New("each port mapping should have a uniquely identified name")
		}

		if v.HealthCheck != nil && !v.HealthCheck.IsEmpty() {
			var (
				protocol = strings.ToLower(v.HealthCheck.Protocol)
				portName = v.HealthCheck.PortName
				path     = strings.TrimSpace(v.HealthCheck.Path)
				command  = strings.TrimSpace(v.HealthCheck.Command)
			)
			if protocol != "cmd" && !utils.SliceContains(portNames, portName) {
				return fmt.Errorf("portname in healthCheck section should match that defined in portMappings")
			}

			if !utils.SliceContains([]string{"tcp", "http", "TCP", "HTTP", "cmd", "CMD"}, protocol) {
				return fmt.Errorf("doesn't recoginized protocol %s for health check", protocol)
			}

			if protocol == "http" {
				if path == "" {
					return fmt.Errorf("no path provided for health check with %s protocol", protocol)
				}
			}

			if protocol == "cmd" {
				if command == "" {
					return fmt.Errorf("no cmd provided for health check with %s protocol", protocol)
				}
			}
		}
	}

	return nil
}
