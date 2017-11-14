package types

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/Dataman-Cloud/swan/utils"
	"github.com/Dataman-Cloud/swan/utils/dfs"
	"github.com/docker/go-connections/nat"
)

var DefaultDeployConfig = DeployConfig{
	WaitDelay: 1,
	Replicas:  1,
}

// ComposeApp sorter
type ComposeAppSorter []*ComposeApp

func (s ComposeAppSorter) Len() int           { return len(s) }
func (s ComposeAppSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ComposeAppSorter) Less(i, j int) bool { return s[i].UpdatedAt.After(s[j].UpdatedAt) }

// wrap ComposeApp with related apps
type ComposeAppWrapper struct {
	*ComposeApp
	Apps []*Application `json:"apps"`
}

// ComposeApp represents a db compose app
type ComposeApp struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	RunAs       string    `json:"run_as"`
	Cluster     string    `json:"cluster"`
	DisplayName string    `json:"display_name"`
	Desc        string    `json:"desc"`
	OpStatus    string    `json:"op_status"` // op status
	ErrMsg      string    `json:"errmsg"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Extra Labels
	Labels map[string]string `json:"labels"`

	// YAML request data
	YAMLRaw string            `json:"yaml_raw"`
	YAMLEnv map[string]string `json:"yaml_env"`

	// held temporary struct convert from YAMLRaw & YAMLEnv and will be converted to App Version
	ComposeV3 *ComposeV3
}

func (cmpApp *ComposeApp) Valid() error {
	if cmpApp.Name == "" {
		return errors.New("compose name required")
	}
	if cmpApp.RunAs == "" {
		return errors.New("compose runas required")
	}
	if cmpApp.Cluster == "" {
		return errors.New("compose cluster required")
	}
	if err := utils.LegalDomain(cmpApp.Name); err != nil {
		return err
	}
	if cmpApp.Name == "default" {
		return errors.New("compose name `default` is reserved")
	}
	return cmpApp.ComposeV3.Valid()
}

func ParseComposeV3(data []byte, envMap map[string]string) (*ComposeV3, error) {
	// first use a RawServiceMap to receive yaml bytes
	var rawComposeV3 = new(RawComposeV3)
	if err := yaml.Unmarshal(data, &rawComposeV3); err != nil {
		return nil, err
	}

	// recursive replace all variables
	if err := InterpolateRawServiceMap(&rawComposeV3.Services, envMap); err != nil {
		return nil, err
	}

	// converted RawServiceMap to expected struct object
	var composeV3 *ComposeV3
	if err := Convert(rawComposeV3, &composeV3); err != nil {
		return nil, err
	}

	// set parsed variable definations
	composeV3.Variables = utils.YamlVariables(data)

	return composeV3, nil
}

// Convert converts a struct (src) to another one (target) using yaml marshalling/unmarshalling.
// If the structure are not compatible, this will throw an error as the unmarshalling will fail.
func Convert(src, target interface{}) error {
	newBytes, err := yaml.Marshal(src)
	if err != nil {
		return fmt.Errorf("Convert.Marshal() yaml error: %v", err)
	}

	err = yaml.Unmarshal(newBytes, target)
	if err != nil {
		return fmt.Errorf("Convert.Unmarshal() yaml error: %v", err)
	}
	return err
}

func (cmpApp *ComposeApp) ParseComposeToVersions() (map[string]*Version, error) {
	cmp := cmpApp.ComposeV3
	if cmp == nil {
		return nil, errors.New("not parsed yet")
	}

	var (
		m       = map[string]*Version{}
		cmpName = cmpApp.Name
		runAs   = cmpApp.RunAs
		cluster = cmpApp.Cluster
		labels  = cmpApp.Labels // Extra Labels
	)

	for svrName := range cmp.Services {
		ver, err := cmp.ConvertServiceToVersion(svrName, cmpName, runAs, cluster, labels)
		if err != nil {
			return nil, fmt.Errorf("convert [%s] error: %v", svrName, err)
		}
		m[svrName] = ver
	}

	return m, nil
}

// Raw Definations to receive raw YAML bytes
//

// RawComposeV3 represent a ComposeV3 struct unparsed
type RawComposeV3 struct {
	Version  string        `yaml:"version"`
	Services RawServiceMap `yaml:"services"`
}

// RawServiceMap a collection of RawServices
type RawServiceMap map[string]RawService

// RawService represent a Service in map form unparsed
type RawService map[string]interface{}

// InterpolateRawServiceMap replaces varialbse in raw service map struct based on environment lookup
func InterpolateRawServiceMap(baseRawServices *RawServiceMap, envMap map[string]string) error {
	for k, v := range *baseRawServices {
		for k2, v2 := range v {
			if err := utils.Interpolate(k2, &v2, envMap); err != nil {
				return err
			}
			(*baseRawServices)[k][k2] = v2
		}
	}
	return nil
}

// ComposeV3 represents parsed docker compose v3 object
// Services (-> map[name]App Version)
// note: do NOT support naming volumes & networks currently
type ComposeV3 struct {
	Version   string                     `yaml:"version"`
	Variables []string                   `yaml:"-"`
	Services  map[string]*ComposeService `yaml:"services"`
}

type ComposeV3Alias ComposeV3 // prevent oom

func (c *ComposeV3) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var alias *ComposeV3Alias
	if err := unmarshal(&alias); err != nil {
		return err
	}
	*c = ComposeV3(*alias)

	for name, svr := range c.Services {
		svr.Name = name
	}
	return nil
}

func (c *ComposeV3) GetServices() []string {
	var srvs = []string{}
	for name := range c.Services {
		srvs = append(srvs, name)
	}
	return srvs
}

func (c *ComposeV3) GetVariables() []string {
	return c.Variables
}

func (c *ComposeV3) Valid() error {
	if len(c.Services) == 0 {
		return errors.New("empty compose service definations")
	}

	// check dependency circled or not
	if err := c.Circled(); err != nil {
		return fmt.Errorf("compose services dependency circled: %v", err)
	}

	// verify each compose service
	for name, srv := range c.Services {
		if name == "" {
			return errors.New("service name required")
		}
		if err := utils.LegalDomain(name); err != nil {
			return err
		}
		if err := srv.Valid(); err != nil {
			return fmt.Errorf("validate service %s error: %v", name, err)
		}
	}

	// ensure uniq on proxy's alias & listen settings
	seenAlias, seenListen := map[string]bool{}, map[string]bool{}
	for name, srv := range c.Services {
		if p := srv.Proxy; p != nil && p.Enabled {
			var (
				alias  = p.Alias
				listen = p.Listen
			)

			if alias != "" {
				if _, ok := seenAlias[alias]; ok {
					return fmt.Errorf("%s proxy alias %s conflict", name, alias)
				}
				seenAlias[alias] = true
			}

			if listen != "" {
				if _, ok := seenListen[listen]; ok {
					return fmt.Errorf("%s proxy listen %s conflict", name, listen)
				}
				seenListen[listen] = true
			}
		}
	}

	return nil
}

func (c *ComposeV3) Circled() error {
	m, err := c.DependMap()
	if err != nil {
		return err
	}
	cs := dfs.NewDirectedCycle(m).Cycle()
	if len(cs) > 0 {
		return fmt.Errorf("dependency circled: %v", cs)
	}
	return nil
}

func (c *ComposeV3) PrioritySort() ([]string, error) {
	m, err := c.DependMap()
	if err != nil {
		return nil, err
	}
	o := dfs.NewDfsOrder(m)
	return o.PostOrder(), nil
}

func (c *ComposeV3) DependMap() (map[string][]string, error) {
	ret := make(map[string][]string)
	for name, svr := range c.Services {
		// ensure exists
		for _, d := range svr.DependsOn {
			if _, ok := c.Services[d]; !ok {
				return nil, fmt.Errorf("missing dependency: %s -> %s", name, d)
			}
		}
		ret[name] = svr.DependsOn
	}
	return ret, nil
}

// convert specified Compose Service to App Version
func (c *ComposeV3) ConvertServiceToVersion(svrName, cmpName, runAs, cluster string, extLabels map[string]string) (*Version, error) {
	svr, ok := c.Services[svrName]
	if !ok {
		return nil, fmt.Errorf("no such compose service: %s", svrName)
	}

	deploy := svr.Deploy
	if deploy == nil {
		deploy = &DefaultDeployConfig
	}

	ver := &Version{
		ID:          fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
		Name:        svrName,
		Env:         svr.Environment,
		Constraints: deploy.Constraints,
		RunAs:       runAs,
		Cluster:     cluster,
		URIs:        svr.URIs,
		IPs:         svr.IPs,
		HealthCheck: svr.healthCheck(),
	}

	dnsSearch := fmt.Sprintf("%s.%s.%s.%s", cmpName, runAs, cluster, swanDomain)

	var err error

	// container
	ver.Container, err = svr.container(dnsSearch, cmpName, extLabels)
	if err != nil {
		return nil, err
	}

	// labels
	lbs := make(map[string]string)
	for k, v := range svr.Labels {
		lbs[k] = v
	}
	for k, v := range extLabels {
		lbs[k] = v
	}
	lbs["SWAN_COMPOSE_NAME"] = cmpName
	ver.Labels = lbs

	// resouces
	if res := svr.Resource; res != nil {
		ver.CPUs, ver.GPUs, ver.Mem, ver.Disk = res.CPUs, res.GPUs, res.Mem, res.Disk
	}

	// command
	if cmd := svr.Command; len(cmd) > 0 {
		ver.Command = strings.Join(cmd, " ")
	}

	// instances
	ver.Instances = int32(deploy.Replicas)

	// killpolicy
	if p := svr.StopGracePeriod; p != "" {
		dur, err := time.ParseDuration(p)
		if err != nil {
			return nil, err
		}
		ver.KillPolicy = &KillPolicy{
			Duration: int64(dur.Seconds()),
		}
	}

	// proxy, just by pass
	ver.Proxy = svr.Proxy

	// validate
	if err := ver.Validate(); err != nil {
		return nil, err
	}
	return ver, nil
}

// ComposeService (-> App Version)
//
//
type ComposeService struct {
	Name            string             `yaml:"name"`
	CapAdd          StrSlice           `yaml:"cap_add"`
	CapDrop         StrSlice           `yaml:"cap_drop"`
	CgroupParent    string             `yaml:"cgroup_parent"`
	Command         StrSlice           `yaml:"command"`
	ContainerName   string             `yaml:"container_name"`
	DependsOn       StrSlice           `yaml:"depends_on"`
	Devices         StrSlice           `yaml:"devices"`
	Dns             StrSlice           `yaml:"dns"`
	DnsSearch       StrSlice           `yaml:"dns_search"`
	DomainName      string             `yaml:"domainname"`
	Entrypoint      StrSlice           `yaml:"entrypoint"`
	Environment     SliceMap           `yaml:"environment"`
	Expose          StrSlice           `yaml:"expose"`
	ExternalLinks   StrSlice           `yaml:"external_links"`
	ExtraHosts      SliceMap           `yaml:"extra_hosts"`
	Hostname        string             `yaml:"hostname"`
	HealthCheck     *HealthCheckConfig `yaml:"healthcheck"`
	Image           string             `yaml:"image"`
	Ipc             string             `yaml:"ipc"`
	Labels          SliceMap           `yaml:"labels"`
	Links           StrSlice           `yaml:"links"`
	Logging         *LoggingConfig     `yaml:"logging"`
	MacAddress      string             `yaml:"mac_address"`
	NetworkMode     string             `yaml:"network_mode"`
	Pid             string             `yaml:"pid"`
	Ports           StrSlice           `yaml:"ports"`
	Privileged      bool               `yaml:"privileged"`
	ReadOnly        bool               `yaml:"read_only"`
	Restart         string             `yaml:"restart"`
	SecurityOpt     StrSlice           `yaml:"security_opt"`
	StdinOpen       bool               `yaml:"stdin_open"`
	StopGracePeriod string             `yaml:"stop_grace_period"`
	StopSignal      string             `yaml:"stop_signal"`
	ShmSize         string             `yaml:"shm_size"`
	Tmpfs           StrSlice           `yaml:"tmpfs"`
	Tty             bool               `yaml:"tty"`
	Ulimits         map[string]*Ulimit `yaml:"ulimits"`
	User            string             `yaml:"user"`
	Volumes         StrSlice           `yaml:"volumes"`
	WorkingDir      string             `yaml:"working_dir"`

	// swan extended attributes
	Deploy     *DeployConfig `yaml:"deploy"`
	Resource   *Resource     `yaml:"resource"`
	PullAlways bool          `yaml:"pull_always"`
	Proxy      *Proxy        `yaml:"proxy"`
	URIs       StrSlice      `yaml:"uris"`
	IPs        StrSlice      `yaml:"ips"`

	// do NOT support attributes
	// Networks        StrSlice           `yaml:"networks"`
}

func (s *ComposeService) Valid() error {
	return utils.LegalDomain(s.Name)
}

func (s *ComposeService) healthCheck() *HealthCheck {
	hc := s.HealthCheck
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
	if t, err := time.ParseDuration(hc.Timeout); err == nil {
		ret.TimeoutSeconds = t.Seconds()
	}
	if t, err := time.ParseDuration(hc.Interval); err == nil {
		ret.IntervalSeconds = t.Seconds()
	}
	ret.ConsecutiveFailures = uint32(hc.Retries)

	return ret
}

func (s *ComposeService) container(dnsSearch, cName string, extLabels map[string]string) (*Container, error) {
	var (
		network    = strings.ToLower(s.NetworkMode)
		image      = s.Image
		forcePull  = s.PullAlways
		privileged = s.Privileged
		parameters = s.parameters(dnsSearch, cName, extLabels)
	)

	portMap, err := s.portMappings(network)
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

// mesos's default supportting for docker container options is so lazy tricky,
// so we have to convert docker container configs to CLI key-value parameter pairs.
func (s *ComposeService) parameters(dnsSearch, cName string, extLabels map[string]string) []*Parameter {
	var (
		m1 = make(map[string]string)   // key-value  params
		m2 = make(map[string][]string) // key-list params
	)

	if v := s.ContainerName; v != "" {
		m1["name"] = v
	}
	if v := s.CgroupParent; v != "" {
		m1["cgroup-parent"] = v
	}
	if v := s.Hostname; v != "" {
		m1["hostname"] = v
	}
	if v := s.Ipc; v != "" {
		m1["ipc"] = v
	}
	if v := s.MacAddress; v != "" {
		m1["mac-address"] = v
	}
	if v := s.Pid; v != "" {
		m1["pid"] = v
	}
	if v := s.StopSignal; v != "" {
		m1["stop-signal"] = v
	}
	if v := s.Restart; v != "" {
		m1["restart"] = v
	}
	if v := s.User; v != "" {
		m1["user"] = v
	}
	if v := s.WorkingDir; v != "" {
		m1["workdir"] = v
	}
	m1["read-only"] = fmt.Sprintf("%t", s.ReadOnly)
	m1["tty"] = fmt.Sprintf("%t", s.Tty)
	// entrypoint
	var e string
	for _, v := range s.Entrypoint {
		e += " " + v
	}
	if e != "" {
		m1["entrypoint"] = e
	}
	// log driver
	if v := s.Logging; v != nil {
		if d := v.Driver; d != "" {
			m1["log-driver"] = d
		}
	}

	// m2
	fset := func(k string, vs []string) {
		m2[k] = append(m2[k], vs...)
	}
	// log-opt
	if v := s.Logging; v != nil {
		opts := make([]string, 0, 0)
		for key, val := range v.Options {
			opts = append(opts, key+"="+val)
		}
		if len(opts) > 0 {
			fset("log-opt", opts)
		}
	}
	if v := s.CapAdd; len(v) > 0 {
		fset("cap-add", v)
	}
	if v := s.CapDrop; len(v) > 0 {
		fset("cap-drop", v)
	}
	if v := s.Devices; len(v) > 0 {
		fset("device", v)
	}
	if v := s.Dns; len(v) > 0 {
		fset("dns", v)
	}
	fset("dns-search", append([]string{dnsSearch}, s.DnsSearch...))

	// env
	if v := s.Environment; len(v) > 0 {
		envs := make([]string, 0, len(v))
		for key, val := range v {
			envs = append(envs, fmt.Sprintf("%s=%s", key, val))
		}
		fset("env", envs)
	}
	// add-host
	if v := s.ExtraHosts; len(v) > 0 {
		hosts := make([]string, 0, len(v))
		for key, val := range v {
			hosts = append(hosts, fmt.Sprintf("%s:%s", key, val))
		}
		fset("add-host", hosts)
	}
	// expose
	if v := s.Expose; len(v) > 0 {
		fset("expose", v)
	}
	if v := s.SecurityOpt; len(v) > 0 {
		fset("security-opt", v)
	}
	// tmpfs
	if v := s.Tmpfs; len(v) > 0 {
		fset("tmpfs", v)
	}
	// labels
	lbs := []string{"SWAN_COMPOSE_NAME=" + cName}
	for key, val := range s.Labels {
		lbs = append(lbs, fmt.Sprintf("%s=%s", key, val))
	}
	for key, val := range extLabels {
		lbs = append(lbs, fmt.Sprintf("%s=%s", key, val))
	}
	fset("label", lbs)
	// volumes
	if v := s.Volumes; len(v) > 0 {
		fset("volume", v)
	}
	// ulimits
	if v := s.Ulimits; len(v) > 0 {
		vs := make([]string, 0, len(v))
		for key, val := range v {
			vs = append(vs, fmt.Sprintf("%s=%d:%d", key, val.Soft, val.Hard))
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

// note: nat.ParsePortSpecs() will make the original port disordered.
// we have to reorder the returned bindings to make fit with mesos port assignments.
func (s *ComposeService) portMappings(network string) ([]*PortMapping, error) {
	var orgPorts []nat.Port
	for _, portSpec := range s.Ports {
		bindings, err := nat.ParsePortSpec(portSpec)
		if err != nil {
			return nil, err
		}
		for _, binding := range bindings {
			orgPorts = append(orgPorts, binding.Port)
		}
	}

	ret := make([]*PortMapping, 0, 0)
	for _, p := range orgPorts {
		cp, _ := strconv.Atoi(p.Port())

		// network bridge set `containerPort` to use.
		if network == "bridge" {
			ret = append(ret, &PortMapping{
				Name:          fmt.Sprintf("%d", cp), // TODO
				ContainerPort: int32(cp),
				Protocol:      strings.ToUpper(p.Proto()),
			})
		}

		// other modes set `hostPort` to use.
		ret = append(ret, &PortMapping{
			Name:     fmt.Sprintf("%d", cp), // TODO
			HostPort: int32(cp),
			Protocol: strings.ToUpper(p.Proto()),
		})

	}

	return ret, nil
}

type LoggingConfig struct {
	Driver  string   `yaml:"driver"`
	Options SliceMap `yaml:"options"`
}

type DeployConfig struct {
	Replicas    int           `yaml:"replicas"`
	WaitDelay   int           `yaml:"wait_delay"`
	Constraints []*Constraint `yaml:"constraints"`
}

type HealthCheckConfig struct {
	Test     StrSlice `yaml:"test"`
	Timeout  string   `yaml:"timeout"`
	Interval string   `yaml:"interval"`
	Retries  int      `yaml:"retries"`
	Disable  bool     `yaml:"disable"`
}

// YAML Complex Type Definations and yaml.Unmarshaler implements
//

// Ulimit represents a single int or a Ulimit
type Ulimit struct {
	Soft int `yaml:"soft"`
	Hard int `yaml:"hard"`
}

type UlimitAlias Ulimit // prevent oom

func (u *Ulimit) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// try single int
	var val int
	if err := unmarshal(&val); err == nil {
		u.Soft = val
		u.Hard = val
		return nil
	}

	// try ulimit
	var alias *UlimitAlias
	if err := unmarshal(&alias); err != nil {
		return err
	}

	*u = Ulimit(*alias)
	return nil
}

// StrSlice represents a single string or a []string
type StrSlice []string

func (s *StrSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// try string
	var cmd string
	if err := unmarshal(&cmd); err == nil {
		*s = StrSlice([]string{cmd})
		return nil
	}

	// try []string
	var cmds []string
	if err := unmarshal(&cmds); err != nil {
		return err
	}
	*s = StrSlice(cmds)
	return nil
}

// StrSlice represents a map[string]strig or a []string
type SliceMap map[string]string

func (m *SliceMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*m = SliceMap(make(map[string]string))

	// try []string
	var ss []string
	if err := unmarshal(&ss); err == nil {
		for _, s := range ss {
			var (
				idxEqual = strings.Index(s, "=")
				idxColon = strings.Index(s, ":")
				splitBy  string
			)
			switch {
			case idxEqual < 0 && idxColon < 0:
				return fmt.Errorf("invalid kv pair: %s, without seperator", s)
			case idxEqual >= 0 && idxColon >= 0:
				if idxEqual < idxColon {
					splitBy = "="
				} else {
					splitBy = ":"
				}
			case idxEqual >= 0:
				splitBy = "="
			case idxColon >= 0:
				splitBy = ":"
			}

			kvpair := strings.SplitN(s, splitBy, 2)
			if len(kvpair) < 2 {
				return fmt.Errorf("invalid key-val pair: %s", s)
			}
			map[string]string(*m)[kvpair[0]] = kvpair[1]
		}
		return nil
	}

	// try map[string]string
	var ms map[string]string
	if err := unmarshal(&ms); err != nil {
		return err
	}

	*m = SliceMap(ms)
	return nil
}
