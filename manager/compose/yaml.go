package compose

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/docker/go-connections/nat"

	"github.com/aanand/compose-file/loader"
	ctypes "github.com/aanand/compose-file/types"
)

var swanDomain string

func init() {
	swanDomain = "swan.com"
	if d := strings.TrimSpace(os.Getenv("SWAN_DOMAIN")); d != "" {
		swanDomain = d
	}
}

// YamlToServiceGroup provide ability to convert docker-compose-yaml to docker-container-config
func YamlToServiceGroup(yaml []byte, env map[string]string, exts map[string]*types.YamlExtra) (types.ServiceGroup, error) {
	cfg, err := YamlServices(yaml, env)
	if err != nil {
		return nil, err
	}

	var (
		ret      = make(map[string]*types.DockerService)
		services = cfg.Services
		networks = cfg.Networks
		volumes  = cfg.Volumes // named volume definations
	)
	for _, svr := range services {
		name := svr.Name

		// extra
		ext, _ := exts[name]
		if ext == nil {
			return nil, errors.New("extra settings requried for service: " + name)
		}

		// service, with extra labels
		nsvr := svr
		if nsvr.Labels == nil {
			nsvr.Labels = make(map[string]string)
		}
		for k, v := range ext.Labels {
			nsvr.Labels[k] = v
		}
		ds := &types.DockerService{
			Name:    name,
			Service: &nsvr,
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

func YamlServices(yaml []byte, env map[string]string) (*ctypes.Config, error) {
	dict, err := loader.ParseYAML(yaml)
	if err != nil {
		return nil, err
	}

	cds := ctypes.ConfigDetails{
		ConfigFiles: []ctypes.ConfigFile{
			{Config: dict},
		},
		Environment: env,
	}

	return loader.Load(cds)
}

// YamlVariables provide ability to parse all of shell variables like:
// $VARIABLE, ${VARIABLE}, ${VARIABLE:-default}, ${VARIABLE-default}
func YamlVariables(yaml []byte) []string {
	var (
		delimiter     = "\\$"
		substitution  = "[_a-z][_a-z0-9]*(?::?-[^}]+)?"
		patternString = fmt.Sprintf(
			"%s(?i:(?P<escaped>%s)|(?P<named>%s)|{(?P<braced>%s)}|(?P<invalid>))",
			delimiter, delimiter, substitution, substitution,
		)
		pattern = regexp.MustCompile(patternString)

		ret = make([]string, 0, 0)
	)

	pattern.ReplaceAllStringFunc(string(yaml), func(sub string) string {
		matches := pattern.FindStringSubmatch(sub)

		groups := make(map[string]string) // all matched naming parts
		for i, name := range pattern.SubexpNames() {
			if i != 0 {
				groups[name] = matches[i]
			}
		}

		text := groups["named"]
		if text == "" {
			text = groups["braced"]
		}
		if text == "" {
			text = groups["escaped"]
		}

		var (
			sep    string
			fields []string
		)
		switch {
		case text == "":
			goto END
		case strings.Contains(text, ":-"):
			sep = ":-"
		case strings.Contains(text, "-"):
			sep = "-"
		default:
			ret = append(ret, text+":")
			goto END
		}

		fields = strings.SplitN(text, sep, 2)
		ret = append(ret, fields[0]+":"+fields[1])

	END:
		return ""
	})

	return ret
}

func SvrToVersion(s *types.DockerService, insName, clusterName string) (*types.Version, error) {
	ver := &types.Version{
		AppName:      s.Name, // svr name
		Priority:     0,      // no use
		Env:          s.Service.Environment,
		Constraints:  s.Extra.Constraints,
		RunAs:        s.Extra.RunAs,
		URIs:         s.Extra.URIs,
		IP:           s.Extra.IPs,
		HealthCheck:  svrToHealthCheck(s),
		UpdatePolicy: nil, // no use
	}

	dnsSearch := fmt.Sprintf("%s.%s.%s.%s", insName, ver.RunAs, clusterName, swanDomain)

	// container
	container, err := svrToContainer(s, dnsSearch, insName)
	if err != nil {
		return nil, err
	}
	ver.Container = container

	// labels
	lbs := make(map[string]string)
	for k, v := range s.Service.Labels {
		lbs[k] = v
	}
	lbs["DM_INSTANCE_NAME"] = insName
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
		ver.KillPolicy = &types.KillPolicy{
			Duration: int64(p.Seconds()),
		}
	}

<<<<<<< HEAD:src/manager/compose/yaml.go
	// gateway
	ver.Gateway = &types.Gateway{
		Enabled: s.Extra.GatewayEnabled,
		Weight:  100,
	}

	return ver, state.ValidateAndFormatVersion(ver)
=======
	return ver, nil
>>>>>>> refactor:manager/compose/yaml.go
}

func svrToHealthCheck(s *types.DockerService) *types.HealthCheck {
	hc := s.Service.HealthCheck
	if hc == nil || hc.Disable {
		return nil
	}

	ret := &types.HealthCheck{
		Protocol: "cmd",
	}

	if t := hc.Test; len(t) > 0 {
		if t[0] == "CMD" || t[0] == "CMD-SHELL" {
			t = t[1:]
		}
		//ret.Value = strings.Join(t, " ")
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

func svrToContainer(s *types.DockerService, dnsSearch, insName string) (*types.Container, error) {
	var (
		network    = strings.ToLower(s.Service.NetworkMode)
		image      = s.Service.Image
		forcePull  = s.Extra.PullAlways
		privileged = s.Service.Privileged
		parameters = svrToParams(s, dnsSearch, insName)
	)
	portMap, err := svrToPortMaps(s)
	if err != nil {
		return nil, err
	}

	return &types.Container{
		Type:    "docker",
		Volumes: nil, // no need, we have convert it to parameters
		Docker: &types.Docker{
			ForcePullImage: forcePull,
			Image:          image,
			Network:        network,
			Parameters:     parameters,
			PortMappings:   portMap,
			Privileged:     privileged,
		},
	}, nil
}

func svrToPortMaps(s *types.DockerService) ([]*types.PortMapping, error) {
	_, binding, err := nat.ParsePortSpecs(s.Service.Ports)
	if err != nil {
		return nil, err
	}

	ret := make([]*types.PortMapping, 0, 0)
	for k := range binding {
		cp, _ := strconv.Atoi(k.Port())
		ret = append(ret, &types.PortMapping{
			Name:          fmt.Sprintf("%d", cp), // TODO
			ContainerPort: int32(cp),
			Protocol:      strings.ToUpper(k.Proto()),
		})
	}
	return ret, nil
}

// sigh ...
// mesos's default supportting for container options is so lazy tricky, so
// we have to convert docker container configs to CLI params key-value pairs.
func svrToParams(s *types.DockerService, dnsSearch, insName string) []*types.Parameter {
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
	lbs := []string{"DM_INSTANCE_NAME=" + insName}
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
	ret := make([]*types.Parameter, 0, 0)
	for k, v := range m1 {
		ret = append(ret, &types.Parameter{k, v})
	}
	for k, vs := range m2 {
		for _, v := range vs {
			ret = append(ret, &types.Parameter{k, v})
		}
	}

	return ret
}
