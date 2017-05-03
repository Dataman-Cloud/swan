package compose

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Dataman-Cloud/swan/src/manager/state"
	"github.com/Dataman-Cloud/swan/src/manager/store"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/aanand/compose-file/loader"
	ctypes "github.com/aanand/compose-file/types"
	uuid "github.com/satori/go.uuid"
)

// YamlToServiceGroup provide ability to convert docker-compose-yaml to docker-container-config
func YamlToServiceGroup(yaml []byte, exts map[string]*store.YamlExtra) (store.ServiceGroup, error) {
	dict, err := loader.ParseYAML(yaml)
	if err != nil {
		return nil, err
	}

	cds := ctypes.ConfigDetails{
		ConfigFiles: []ctypes.ConfigFile{
			{Config: dict},
		},
		Environment: map[string]string{}, // TODO
	}

	cfg, err := loader.Load(cds)
	if err != nil {
		return nil, err
	}

	var (
		ret      = make(map[string]*store.DockerService)
		services = cfg.Services
		networks = cfg.Networks
		volumes  = cfg.Volumes
	)
	for _, svr := range services {
		// service
		name := svr.Name
		nsvr := svr
		ds := &store.DockerService{
			Name:    name,
			Service: &nsvr,
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
		// extra
		ext, _ := exts[name]
		if ext == nil {
			return nil, errors.New("extra settings requried for service: " + name)
		}
		ds.Extra = ext

		ret[name] = ds
	}

	return ret, nil
}

func SvrToVersion(s *store.DockerService, insName string) (*types.Version, error) {
	ver := &types.Version{
		ID:           uuid.NewV4().String(),
		AppName:      s.Name, // svr name
		AppVersion:   "v1",   // default
		Priority:     0,      // no use
		Env:          s.Service.Environment,
		Constraints:  s.Extra.Constraints,
		RunAs:        s.Extra.RunAs,
		URIs:         s.Extra.URIs,
		IP:           s.Extra.IPs,
		Container:    svrToContainer(s),
		HealthCheck:  nil, // not supported yet
		UpdatePolicy: nil, // no use
	}

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
	case "global": // exactly one container per swarm node
		ver.Constraints = "unique hostname" // overwrite constraints
		ver.Instances = 1                   // no use, just to make pass validator
	}

	// killpolicy
	if p := s.Service.StopGracePeriod; p != nil {
		ver.KillPolicy = &types.KillPolicy{
			Duration: int64(p.Seconds()),
		}
	}

	return ver, state.ValidateAndFormatVersion(ver)
}

func svrToContainer(s *store.DockerService) *types.Container {
	// network
	var (
		network    = strings.ToLower(s.Service.NetworkMode)
		image      = s.Service.Image
		forcePull  = s.Extra.PullAlways
		privileged = s.Service.Privileged
		parameters = svrToParams(s)
	)

	// ports

	// volumes
	volumes := make([]*types.Volume, 0, 0)

	return &types.Container{
		Type: "docker",
		Docker: &types.Docker{
			ForcePullImage: forcePull,
			Image:          image,
			Network:        network,
			Parameters:     parameters,
			PortMappings:   nil, // TODO
			Privileged:     privileged,
		},
		Volumes: volumes,
	}
}

// sigh ...
// mesos's default supportting for container options is so lazy tricky, so
// we have to convert docker container configs to CLI params key-value pairs.
func svrToParams(s *store.DockerService) []*types.Parameter {
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
	if v := s.Service.DnsSearch; len(v) > 0 {
		fset("dns-search", v)
	}

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
	// TODO need rewrite links
	//if v := s.Service.Links; len(v) > 0 {
	//fset("link", v)
	//}
	if v := s.Service.SecurityOpt; len(v) > 0 {
		fset("security-opt", v)
	}
	// tmpfs
	if v := s.Service.Tmpfs; len(v) > 0 {
		fset("tmpfs", v)
	}
	// labels
	if v := s.Service.Labels; len(v) > 0 {
		lbs := make([]string, 0, len(v))
		for key, val := range v {
			lbs = append(lbs, fmt.Sprintf("%s=%s", key, val))
		}
		fset("label", lbs)
	}
	// volumes
	if v := s.Service.Volumes; len(v) > 0 {
		fset("volume", v)
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
