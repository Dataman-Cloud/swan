package types

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/utils"
	"github.com/Dataman-Cloud/swan/utils/dfs"
	"github.com/aanand/compose-file/types"
	"github.com/docker/go-connections/nat"
)

var swanDomain string

func init() {
	swanDomain = "swan.com"
	if d := strings.TrimSpace(os.Getenv("SWAN_DOMAIN")); d != "" {
		swanDomain = d
	}
}

// save to -> keyCompose
type Compose struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Desc        string    `json:"desc"`
	Status      string    `json:"status"` // op status
	ErrMsg      string    `json:"errmsg"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// request settings
	ServiceGroup ServiceGroup          `json:"service_group"`
	YAMLRaw      string                `json:"yaml_raw"`
	YAMLEnv      map[string]string     `json:"yaml_env"`
	YAMLExtra    map[string]*YamlExtra `json:"yaml_extra"`
}

func (c *Compose) RequireConvert() bool {
	return len(c.ServiceGroup) == 0 && c.YAMLRaw != ""
}

func (c *Compose) Valid() error {
	reg := regexp.MustCompile(`^[a-zA-Z0-9]{1,32}$`)
	if !reg.MatchString(c.Name) {
		return errors.New("instance name should be regexp matched by: " + reg.String())
	}
	if c.Name == "default" {
		return errors.New("instance name reserved")
	}

	return c.ServiceGroup.Valid()
}

func (c *Compose) ToServiceGroup() (ServiceGroup, error) {
	cfg, err := utils.YamlServices([]byte(c.YAMLRaw), c.YAMLEnv)
	if err != nil {
		return nil, err
	}

	var (
		ret      = make(map[string]*DockerService)
		services = cfg.Services
		networks = cfg.Networks
		volumes  = cfg.Volumes // named volume definations
	)
	for _, srv := range services {
		name := srv.Name

		// extra
		ext, _ := c.YAMLExtra[name]
		if ext == nil {
			return nil, errors.New("extra settings requried for service: " + name)
		}

		// service, with extra labels
		nsrv := srv
		if nsrv.Labels == nil {
			nsrv.Labels = make(map[string]string)
		}
		for k, v := range ext.Labels {
			nsrv.Labels[k] = v
		}
		ds := &DockerService{
			Name:    name,
			Service: &nsrv,
			Extra:   ext,
		}

		// network
		if v, ok := networks[name]; ok {
			nv := v
			ds.Network = &nv
		}

		// volume
		if v, ok := volumes[name]; ok {
			nv := v
			ds.Volume = &nv
		}

		ret[name] = ds
	}

	return ret, nil
}

type Resource struct {
	CPU   float64  `json:"cpu"`
	Mem   float64  `json:"mem"`
	Disk  float64  `json:"disk"`
	Ports []uint64 `json:"ports"`
}

type YamlExtra struct {
	Priority    uint              `json:"priority"`
	WaitDelay   uint              `json:"wait_delay"` // by second
	PullAlways  bool              `json:"pull_always"`
	Resource    *Resource         `json:"resource"`
	Constraints []*Constraint     `json:"constraints"`
	RunAs       string            `json:"runas"`
	URIs        []string          `json:"uris"`
	IPs         []string          `json:"ips"`
	Labels      map[string]string `json:"labels"` // extra labels: uid, username, vcluster ...
	Proxy       *Proxy            `json:"proxy"`
}

type ServiceGroup map[string]*DockerService

func (sg ServiceGroup) Valid() error {
	if len(sg) == 0 {
		return errors.New("serviceGroup empty")
	}
	for name, srv := range sg {
		if name == "" {
			return errors.New("service name required")
		}
		if strings.ContainsRune(name, '-') {
			return errors.New(`char '-' not allowed for service name`)
		}
		if name != srv.Name {
			return errors.New("service name mismatched")
		}
		if err := srv.Valid(); err != nil {
			return fmt.Errorf("validate service %s error: %v", name, err)
		}
	}
	return sg.circled()
}

func (sg ServiceGroup) PrioritySort() ([]string, error) {
	m, err := sg.dependMap()
	if err != nil {
		return nil, err
	}
	o := dfs.NewDfsOrder(m)
	return o.PostOrder(), nil
}

func (sg ServiceGroup) circled() error {
	m, err := sg.dependMap()
	if err != nil {
		return err
	}
	c := dfs.NewDirectedCycle(m)
	if cs := c.Cycle(); len(cs) > 0 {
		return fmt.Errorf("dependency circled: %v", cs)
	}
	return nil
}

func (sg ServiceGroup) dependMap() (map[string][]string, error) {
	ret := make(map[string][]string)
	for name, svr := range sg {
		// ensure exists
		for _, d := range svr.Service.DependsOn {
			if _, ok := sg[d]; !ok {
				return nil, fmt.Errorf("missing dependency: %s -> %s", name, d)
			}
		}
		ret[name] = svr.Service.DependsOn
	}
	return ret, nil
}

type DockerService struct {
	Name    string               `json:"name"`
	Service *types.ServiceConfig `json:"service"`
	Network *types.NetworkConfig `json:"network"`
	Volume  *types.VolumeConfig  `json:"volume"`
	Extra   *YamlExtra           `json:"extra"`
}

func (s *DockerService) Valid() error {
	return nil
}

func (s *DockerService) ToVersion(cName, cluster string) (*Version, error) {
	ver := &Version{
		Name:         s.Name, // svr name
		Priority:     0,      // no use
		Env:          s.Service.Environment,
		Constraints:  s.Extra.Constraints,
		RunAs:        s.Extra.RunAs,
		URIs:         s.Extra.URIs,
		IPs:          s.Extra.IPs,
		HealthCheck:  s.healthCheck(),
		UpdatePolicy: nil, // no use
	}

	dnsSearch := fmt.Sprintf("%s.%s.%s.%s", cName, ver.RunAs, cluster, swanDomain)

	// container
	container, err := s.container(dnsSearch, cName)
	if err != nil {
		return nil, err
	}
	ver.Container = container

	// labels
	lbs := make(map[string]string)
	for k, v := range s.Service.Labels {
		lbs[k] = v
	}
	lbs["DM_INSTANCE_NAME"] = cName
	ver.Labels = lbs

	// resouces
	if res := s.Extra.Resource; res != nil {
		ver.CPUs, ver.Mem, ver.Disk = res.CPU, res.Mem, res.Disk
	}

	// command
	if cmd := s.Service.Command; len(cmd) > 0 {
		ver.Command = strings.Join(cmd, " ")
	}

	// instances
	switch m := s.Service.Deploy.Mode; m {
	case "", "replicated": // specified number of containers
		if n := s.Service.Deploy.Replicas; n == nil {
			ver.Instances = int32(1)
		} else {
			ver.Instances = int32(*n)
		}
	default:
		ver.Instances = 1
	}

	// killpolicy
	if p := s.Service.StopGracePeriod; p != nil {
		ver.KillPolicy = &KillPolicy{
			Duration: int64(p.Seconds()),
		}
	}

	// proxy
	ver.Proxy = &Proxy{
		Enabled: s.Extra.Proxy.Enabled,
		Alias:   s.Extra.Proxy.Alias,
	}

	if err := ver.Validate(); err != nil {
		return nil, err
	}

	return ver, nil
}

func (s *DockerService) healthCheck() *HealthCheck {
	hc := s.Service.HealthCheck
	if hc == nil || hc.Disable {
		return nil
	}

	ret := &HealthCheck{
		Protocol: "cmd",
	}

	if t := hc.Test; len(t) > 0 {
		if t[0] == "CMD" || t[0] == "CMD-SHELL" {
			t = t[1:]
		}
		ret.Command = strings.Join(t, " ")
	}
	// Value:    strings.Join(hc.Test, " "),
	if t, err := time.ParseDuration(hc.Timeout); err == nil {
		ret.TimeoutSeconds = t.Seconds()
	}
	if t, err := time.ParseDuration(hc.Interval); err == nil {
		ret.IntervalSeconds = t.Seconds()
	}
	if r := hc.Retries; r != nil {
		ret.ConsecutiveFailures = uint32(*r)
	}

	return ret
}

func (s *DockerService) container(dnsSearch, cName string) (*Container, error) {
	var (
		network    = strings.ToLower(s.Service.NetworkMode)
		image      = s.Service.Image
		forcePull  = s.Extra.PullAlways
		privileged = s.Service.Privileged
		parameters = s.parameters(dnsSearch, cName)
	)
	portMap, err := s.portMappings()
	if err != nil {
		return nil, err
	}

	return &Container{
		Type:    "docker",
		Volumes: nil, // no need, we have convert it to parameters
		Docker: &Docker{
			ForcePullImage: forcePull,
			Image:          image,
			Network:        network,
			Parameters:     parameters,
			PortMappings:   portMap,
			Privileged:     privileged,
		},
	}, nil
}

func (s *DockerService) parameters(dnsSearch, cName string) []*Parameter {
	var (
		m1 = make(map[string]string)   // key-value  params
		m2 = make(map[string][]string) // key-list params
	)

	if v := s.Service.ContainerName; v != "" {
		m1["name"] = v
	}
	if v := s.Service.CgroupParent; v != "" {
		m1["cgroup-parent"] = v
	}
	if v := s.Service.Hostname; v != "" {
		m1["hostname"] = v
	}
	if v := s.Service.Ipc; v != "" {
		m1["ipc"] = v
	}
	if v := s.Service.MacAddress; v != "" {
		m1["mac-address"] = v
	}
	if v := s.Service.Pid; v != "" {
		m1["pid"] = v
	}
	if v := s.Service.StopSignal; v != "" {
		m1["stop-signal"] = v
	}
	if v := s.Service.Restart; v != "" {
		m1["restart"] = v
	}
	if v := s.Service.User; v != "" {
		m1["user"] = v
	}
	if v := s.Service.WorkingDir; v != "" {
		m1["workdir"] = v
	}
	m1["read-only"] = fmt.Sprintf("%t", s.Service.ReadOnly)
	m1["tty"] = fmt.Sprintf("%t", s.Service.Tty)
	// entrypoint
	var e string
	for _, v := range s.Service.Entrypoint {
		e += " " + v
	}
	if e != "" {
		m1["entrypoint"] = e
	}
	// logging
	if v := s.Service.Logging; v != nil {
		if d := v.Driver; d != "" {
			m1["log-driver"] = d
		}
		var opts string
		for key, val := range v.Options {
			if len(opts) > 0 {
				opts += " " + key + "=" + val
			} else {
				opts += key + "=" + val
			}
		}
		if opts != "" {
			m1["log-opt"] = opts
		}
	}

	// m2
	fset := func(k string, vs []string) {
		m2[k] = append(m2[k], vs...)
	}
	if v := s.Service.CapAdd; len(v) > 0 {
		fset("cap-add", v)
	}
	if v := s.Service.CapDrop; len(v) > 0 {
		fset("cap-drop", v)
	}
	if v := s.Service.Devices; len(v) > 0 {
		fset("device", v)
	}
	if v := s.Service.Dns; len(v) > 0 {
		fset("dns", v)
	}
	fset("dns-search", []string{dnsSearch})

	// env
	if v := s.Service.Environment; len(v) > 0 {
		envs := make([]string, 0, len(v))
		for key, val := range v {
			envs = append(envs, fmt.Sprintf("%s=%s", key, val))
		}
		fset("env", envs)
	}
	// add-host
	if v := s.Service.ExtraHosts; len(v) > 0 {
		hosts := make([]string, 0, len(v))
		for key, val := range v {
			hosts = append(hosts, fmt.Sprintf("%s:%s", key, val))
		}
		fset("add-host", hosts)
	}
	// expose
	if v := s.Service.Expose; len(v) > 0 {
		fset("expose", v)
	}
	if v := s.Service.SecurityOpt; len(v) > 0 {
		fset("security-opt", v)
	}
	// tmpfs
	if v := s.Service.Tmpfs; len(v) > 0 {
		fset("tmpfs", v)
	}
	// labels
	lbs := []string{"DM_INSTANCE_NAME=" + cName}
	for key, val := range s.Service.Labels {
		lbs = append(lbs, fmt.Sprintf("%s=%s", key, val))
	}
	fset("label", lbs)
	// volumes
	if v := s.Service.Volumes; len(v) > 0 {
		fset("volume", v)
	}
	// ulimits
	if v := s.Service.Ulimits; len(v) > 0 {
		vs := make([]string, 0, len(v))
		for key, val := range v {
			if val.Single > 0 {
				vs = append(vs, fmt.Sprintf("%s=%d:%d", key, val.Single, val.Single))
			} else {
				vs = append(vs, fmt.Sprintf("%s=%d:%d", key, val.Soft, val.Hard))
			}
		}
		fset("ulimit", vs)
	}
	// final
	ret := make([]*Parameter, 0, 0)
	for k, v := range m1 {
		ret = append(ret, &Parameter{k, v})
	}
	for k, vs := range m2 {
		for _, v := range vs {
			ret = append(ret, &Parameter{k, v})
		}
	}

	return ret
}

func (s *DockerService) portMappings() ([]*PortMapping, error) {
	_, binding, err := nat.ParsePortSpecs(s.Service.Ports)
	if err != nil {
		return nil, err
	}

	ret := make([]*PortMapping, 0, 0)
	for k := range binding {
		cp, _ := strconv.Atoi(k.Port())
		ret = append(ret, &PortMapping{
			Name:          fmt.Sprintf("%d", cp), // TODO
			ContainerPort: int32(cp),
			Protocol:      strings.ToUpper(k.Proto()),
		})
	}

	return ret, nil
}
