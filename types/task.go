package types

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/mesosproto"
	"github.com/gogo/protobuf/proto"
)

const (
	TaskHealthyUnset = "unset"
	TaskHealthy      = "healty"
	TaskUnHealthy    = "unhealthy"
)

type Task struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	IP            string    `json:"ip"`
	Ports         []uint64  `json:"ports"`
	Healthy       string    `json:"healthy"`
	Weight        float64   `json:"weight"`
	AgentId       string    `json:"agentId"`
	Version       string    `json:"version"`
	Status        string    `json:"status"`
	ErrMsg        string    `json:"errmsg"`
	OpStatus      string    `json:"opstatus"`
	ContainerID   string    `json:"container_id"`
	ContainerName string    `json:"container_name"`
	MaxRetries    int       `json:"maxRetries"`
	Histories     []*Task   `json:"histories"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
}

type TaskList []*Task

func (tl TaskList) Len() int      { return len(tl) }
func (tl TaskList) Swap(i, j int) { tl[i], tl[j] = tl[j], tl[i] }
func (tl TaskList) Less(i, j int) bool {
	m, _ := strconv.Atoi(strings.Split(tl[i].Name, ".")[0])
	n, _ := strconv.Atoi(strings.Split(tl[j].Name, ".")[0])

	return m < n
}

func (tl TaskList) Reverse() {
	sort.Sort(sort.Reverse(tl))
}

func (tl TaskList) Sort() {
	sort.Sort(tl)
}

type TaskConfig struct {
	CPUs           float64           `json:"cpus"`
	GPUs           float64           `json:"gpus"`
	Mem            float64           `json:"mem"`
	Disk           float64           `json:"disk"`
	IP             string            `json:"ip"`
	Ports          []uint64          `json:"ports"`
	Image          string            `json:"image"`
	Command        string            `json:"cmd"`
	Privileged     bool              `json:"privileged"`
	ForcePullImage bool              `json:"forcePullImage"`
	Volumes        []*Volume         `json:"volumes"`
	PortMappings   []*PortMapping    `json:"portmappings"`
	Network        string            `json:"network"`
	Parameters     []*Parameter      `json:"parameters"`
	HealthCheck    *HealthCheck      `json:"healthCheck"`
	KillPolicy     *KillPolicy       `json:"killPolicy"`
	RestartPolicy  *RestartPolicy    `json:"restart"`
	Labels         map[string]string `json:"labels"`
	URIs           []string          `json:"uris"`
	Env            map[string]string `json:"env"`
	Constraints    []*Constraint     `json:"constraints"`
	Proxy          *Proxy            `json:"proxy"`
}

func NewTaskConfig(spec *Version, idx int) *TaskConfig {
	cfg := &TaskConfig{
		CPUs:           spec.CPUs,
		GPUs:           spec.GPUs,
		Mem:            spec.Mem,
		Disk:           spec.Disk,
		Image:          spec.Container.Docker.Image,
		Command:        spec.Command,
		Privileged:     spec.Container.Docker.Privileged,
		ForcePullImage: spec.Container.Docker.ForcePullImage,
		Volumes:        spec.Container.Volumes,
		PortMappings:   spec.Container.Docker.PortMappings,
		Network:        spec.Container.Docker.Network,
		Parameters:     spec.Container.Docker.Parameters,
		HealthCheck:    spec.HealthCheck,
		KillPolicy:     spec.KillPolicy,
		RestartPolicy:  spec.RestartPolicy,
		Labels:         spec.Labels,
		URIs:           spec.URIs,
		Env:            spec.Env,
		Constraints:    spec.Constraints,
		Proxy:          spec.Proxy,
	}

	// with user specified ip address
	if cfg.Network != "host" && cfg.Network != "bridge" {
		cfg.Parameters = append(cfg.Parameters, &Parameter{
			Key:   "ip",
			Value: spec.IPs[idx],
		})
		cfg.IP = spec.IPs[idx]
	}

	return cfg
}

func (c *TaskConfig) BuildCommand() *mesosproto.CommandInfo {
	if cmd := c.Command; len(cmd) > 0 {
		return &mesosproto.CommandInfo{
			Uris:        c.uris(),
			Environment: c.envs(),
			Shell:       proto.Bool(true), // sh -c "cmd"
			Value:       proto.String(cmd),
		}
	}

	return &mesosproto.CommandInfo{Shell: proto.Bool(false)}
}

func (c *TaskConfig) uris() []*mesosproto.CommandInfo_URI {
	uris := make([]*mesosproto.CommandInfo_URI, 0)

	for _, v := range c.URIs {
		uris = append(uris, &mesosproto.CommandInfo_URI{
			Value: proto.String(v),
		})
	}

	return uris
}

func (c *TaskConfig) envs() *mesosproto.Environment {
	vars := make([]*mesosproto.Environment_Variable, 0)

	for k, v := range c.Env {
		vars = append(vars, &mesosproto.Environment_Variable{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}

	return &mesosproto.Environment{
		Variables: vars,
	}
}

func (c *TaskConfig) volumes() []*mesosproto.Volume {
	var (
		vlms = c.Volumes
		mvs  = make([]*mesosproto.Volume, 0, 0)
	)

	for _, vlm := range vlms {
		mode := mesosproto.Volume_RO
		if vlm.Mode == "RW" {
			mode = mesosproto.Volume_RW
		}

		mvs = append(mvs, &mesosproto.Volume{
			ContainerPath: proto.String(vlm.ContainerPath),
			HostPath:      proto.String(vlm.HostPath),
			Mode:          &mode,
		})
	}

	return mvs
}

func (c *TaskConfig) network() *mesosproto.ContainerInfo_DockerInfo_Network {
	var (
		network = c.Network
	)

	switch network {
	case "none":
		return mesosproto.ContainerInfo_DockerInfo_NONE.Enum()
	case "host":
		return mesosproto.ContainerInfo_DockerInfo_HOST.Enum()
	case "bridge":
		return mesosproto.ContainerInfo_DockerInfo_BRIDGE.Enum()
	}

	// mesosproto.ContainerInfo_DockerInfo_USER always lead to error complains:
	// Failed to run docker container: No network info found in container info
	// we process user-defined network within parameters().
	return mesosproto.ContainerInfo_DockerInfo_NONE.Enum()
}

func (c *TaskConfig) portMappings() []*mesosproto.ContainerInfo_DockerInfo_PortMapping {
	var (
		pms  = c.PortMappings
		dpms = make([]*mesosproto.ContainerInfo_DockerInfo_PortMapping, 0, 0)
	)

	if c.Network == "bridge" {
		for idx, m := range pms {
			dpms = append(dpms,
				&mesosproto.ContainerInfo_DockerInfo_PortMapping{
					HostPort:      proto.Uint32(uint32(c.Ports[idx])),
					ContainerPort: proto.Uint32(uint32(m.ContainerPort)),
					Protocol:      proto.String(m.Protocol),
				})
		}
	}

	return dpms
}

func (c *TaskConfig) parameters() []*mesosproto.Parameter {
	var (
		ps  = c.Parameters
		mps = make([]*mesosproto.Parameter, 0, 0)
	)

	for _, p := range ps {
		mps = append(mps, &mesosproto.Parameter{
			Key:   proto.String(p.Key),
			Value: proto.String(p.Value),
		})
	}

	// attach with user defined network parameters
	var (
		network = c.Network
		ip      = c.IP
	)
	switch network {
	case "none", "bridge", "host":
	default:
		mps = append(mps, &mesosproto.Parameter{
			Key:   proto.String("network"),
			Value: proto.String(network), // user defined network name
		})

		if ip != "" {
			mps = append(mps, &mesosproto.Parameter{
				Key:   proto.String("ip"),
				Value: proto.String(ip), // user defined ip address
			})
		}
	}

	return mps
}

func (c *TaskConfig) BuildContainer() *mesosproto.ContainerInfo {
	var (
		image      = c.Image
		privileged = c.Privileged
		force      = c.ForcePullImage
	)

	return &mesosproto.ContainerInfo{
		Type:    mesosproto.ContainerInfo_DOCKER.Enum(),
		Volumes: c.volumes(),
		Docker: &mesosproto.ContainerInfo_DockerInfo{
			Image:          proto.String(image),
			Privileged:     proto.Bool(privileged),
			Network:        c.network(),
			PortMappings:   c.portMappings(),
			Parameters:     c.parameters(),
			ForcePullImage: proto.Bool(force),
		},
	}

}

func (c *TaskConfig) BuildResources() []*mesosproto.Resource {
	var (
		rs   = make([]*mesosproto.Resource, 0, 0)
		cpus = c.CPUs
		gpus = c.GPUs
		mem  = c.Mem
		disk = c.Disk
	)

	if cpus > 0 {
		rs = append(rs, &mesosproto.Resource{
			Name: proto.String("cpus"),
			Type: mesosproto.Value_SCALAR.Enum(),
			Scalar: &mesosproto.Value_Scalar{
				Value: proto.Float64(cpus),
			},
		})
	}

	if gpus > 0 {
		rs = append(rs, &mesosproto.Resource{
			Name: proto.String("gpus"),
			Type: mesosproto.Value_SCALAR.Enum(),
			Scalar: &mesosproto.Value_Scalar{
				Value: proto.Float64(gpus),
			},
		})
	}

	if mem > 0 {
		rs = append(rs, &mesosproto.Resource{
			Name: proto.String("mem"),
			Type: mesosproto.Value_SCALAR.Enum(),
			Scalar: &mesosproto.Value_Scalar{
				Value: proto.Float64(mem),
			},
		})
	}

	if disk > 0 {
		rs = append(rs, &mesosproto.Resource{
			Name: proto.String("disk"),
			Type: mesosproto.Value_SCALAR.Enum(),
			Scalar: &mesosproto.Value_Scalar{
				Value: proto.Float64(disk),
			},
		})
	}

	for idx, pm := range c.PortMappings {
		switch c.Network {
		case "host":
			if pm.HostPort == 0 {
				rs = append(rs, &mesosproto.Resource{
					Name: proto.String("ports"),
					Type: mesosproto.Value_RANGES.Enum(),
					Ranges: &mesosproto.Value_Ranges{
						Range: []*mesosproto.Value_Range{
							{
								Begin: proto.Uint64(c.Ports[idx]),
								End:   proto.Uint64(c.Ports[idx]),
							},
						},
					},
				})
			}

			c.Env[fmt.Sprintf("SWAN_HOST_PORT_%s", strings.ToUpper(pm.Name))] = fmt.Sprintf("%d", c.Ports[idx])
		case "bridge":
			rs = append(rs, &mesosproto.Resource{
				Name: proto.String("ports"),
				Type: mesosproto.Value_RANGES.Enum(),
				Ranges: &mesosproto.Value_Ranges{
					Range: []*mesosproto.Value_Range{
						{
							Begin: proto.Uint64(c.Ports[idx]),
							End:   proto.Uint64(c.Ports[idx]),
						},
					},
				},
			})
		}
	}

	return rs
}

func (c *TaskConfig) BuildHealthCheck() *mesosproto.HealthCheck {
	var (
		health *mesosproto.HealthCheck
		port   int32
	)

	network := strings.ToLower(c.Network)
	for _, pm := range c.PortMappings {
		if c.HealthCheck.PortName == pm.Name {
			if network == "host" {
				port = pm.HostPort
			}

			if network == "bridge" {
				port = pm.ContainerPort
			}

			break
		}
	}

	protocol := strings.ToLower(c.HealthCheck.Protocol)
	switch protocol {
	case "cmd":
		health = &mesosproto.HealthCheck{
			Type: mesosproto.HealthCheck_COMMAND.Enum(),
			Command: &mesosproto.CommandInfo{
				Value: proto.String(c.HealthCheck.Command),
			},
		}
	case "http":

		health = &mesosproto.HealthCheck{
			Type: mesosproto.HealthCheck_HTTP.Enum(),
			Http: &mesosproto.HealthCheck_HTTPCheckInfo{
				Scheme:   proto.String(protocol),
				Port:     proto.Uint32(uint32(port)),
				Path:     proto.String(c.HealthCheck.Path),
				Statuses: []uint32{uint32(200), uint32(201), uint32(301), uint32(302)},
			},
		}
	case "tcp":
		health = &mesosproto.HealthCheck{
			Type: mesosproto.HealthCheck_TCP.Enum(),
			Tcp: &mesosproto.HealthCheck_TCPCheckInfo{
				Port: proto.Uint32(uint32(port)),
			},
		}
	}

	health.DelaySeconds = proto.Float64(c.HealthCheck.DelaySeconds)
	health.IntervalSeconds = proto.Float64(c.HealthCheck.IntervalSeconds)
	health.TimeoutSeconds = proto.Float64(c.HealthCheck.TimeoutSeconds)
	health.ConsecutiveFailures = proto.Uint32(c.HealthCheck.ConsecutiveFailures)
	health.GracePeriodSeconds = proto.Float64(c.HealthCheck.GracePeriodSeconds)

	return health
}

func (c *TaskConfig) BuildKillPolicy() *mesosproto.KillPolicy {
	return &mesosproto.KillPolicy{
		GracePeriod: &mesosproto.DurationInfo{
			Nanoseconds: proto.Int64(c.KillPolicy.Duration * 1000 * 1000),
		},
	}
}
func (c *TaskConfig) BuildLabels(name string) *mesosproto.Labels {
	appId := strings.SplitN(name, ".", 2)[1]

	labl := &mesosproto.Label{
		Key:   proto.String("app_name"),
		Value: proto.String(appId),
	}

	labls := make([]*mesosproto.Label, 0)

	labls = append(labls, labl)

	return &mesosproto.Labels{
		Labels: labls,
	}
}

type TaskSorter []*Task

func (s TaskSorter) Len() int      { return len(s) }
func (s TaskSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s TaskSorter) Less(i, j int) bool {
	a, _ := strconv.Atoi(strings.Split(s[i].Name, ".")[0])
	b, _ := strconv.Atoi(strings.Split(s[j].Name, ".")[0])

	return a < b
}

func (t *Task) Index() string {
	return strings.Split(t.Name, ".")[0]
}
