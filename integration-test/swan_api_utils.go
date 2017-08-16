package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"text/template"
	"time"

	"gopkg.in/check.v1"

	"github.com/Dataman-Cloud/swan/types"
)

var (
	errWaitAppTimeout     = errors.New("wait app timeout")
	errWaitComposeTimeout = errors.New("wait compose timeout")
	errWaitPurgeTimeout   = errors.New("wait purge timeout")
)

// wait App status reached to expected status until timeout
func (s *ApiSuite) waitApp(id, expect string, maxWait time.Duration, c *check.C) error {
	timeout := time.After(maxWait)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return errWaitAppTimeout

		case <-ticker.C:
			app := s.inspectApp(id, c)
			if app.OpStatus == expect {
				return nil
			}
			c.Logf("App: %s is %s ...", id, app.OpStatus)
		}
	}
}

func (s *ApiSuite) updateApp(id string, newVer *types.Version, c *check.C) {
	code, body, err := s.sendRequest("PUT", "/v1/apps/"+id, newVer)
	c.Assert(err, check.IsNil)
	c.Log(string(body))
	c.Assert(code, check.Equals, http.StatusAccepted)
}

func (s *ApiSuite) rollbackApp(id string, versionID string, c *check.C) {
	code, body, err := s.rawRollBackApp(id, versionID)
	c.Assert(err, check.IsNil)
	c.Log(string(body))
	c.Assert(code, check.Equals, http.StatusAccepted)
}

func (s *ApiSuite) rawRollBackApp(id string, versionID string) (int, []byte, error) {
	uri := fmt.Sprintf("/v1/apps/%s/rollback?version=%s", id, versionID)
	return s.sendRequest("POST", uri, nil)
}

func (s *ApiSuite) canaryUpdate(id string, canary *types.CanaryUpdateBody, c *check.C) {
	code, body, err := s.sendRequest("PUT", "/v1/apps/"+id+"/canary", canary)
	c.Assert(err, check.IsNil)
	c.Log(string(body))
	c.Assert(code, check.Equals, http.StatusAccepted)
}

func (s *ApiSuite) createApp(ver *types.Version, c *check.C) string {
	code, body, err := s.rawCreateApp(ver)
	c.Assert(err, check.IsNil)
	c.Log(string(body))
	c.Assert(code, check.Equals, http.StatusCreated)

	var resp struct {
		Id string
	}
	err = s.bind(body, &resp)
	c.Assert(err, check.IsNil)
	c.Logf("created App %s", resp.Id)
	c.Assert(resp.Id, check.Matches, ver.Name+".*")

	return resp.Id
}

func (s *ApiSuite) rawCreateApp(ver *types.Version) (int, []byte, error) {
	return s.sendRequest("POST", "/v1/apps", ver)
}

func (s *ApiSuite) listApps(c *check.C) []*types.Application {
	code, body, err := s.sendRequest("GET", "/v1/apps", nil)
	c.Assert(err, check.IsNil)
	c.Log(string(body))
	c.Assert(code, check.Equals, http.StatusOK)

	var apps []*types.Application
	err = s.bind(body, &apps)
	c.Assert(err, check.IsNil)

	return apps
}

func (s *ApiSuite) listAppVersions(id string, c *check.C) []*types.Version {
	code, body, err := s.sendRequest("GET", "/v1/apps/"+id+"/versions", nil)
	c.Assert(err, check.IsNil)
	c.Log(string(body))
	c.Assert(code, check.Equals, http.StatusOK)

	var vers []*types.Version
	err = s.bind(body, &vers)
	c.Assert(err, check.IsNil)

	return vers
}

func (s *ApiSuite) listAppTasks(id string, c *check.C) []*types.Task {
	code, body, err := s.sendRequest("GET", "/v1/apps/"+id+"/tasks", nil)
	c.Assert(err, check.IsNil)
	c.Log(string(body))
	c.Assert(code, check.Equals, http.StatusOK)

	var tasks []*types.Task
	err = s.bind(body, &tasks)
	c.Assert(err, check.IsNil)

	return tasks
}

func (s *ApiSuite) inspectApp(id string, c *check.C) *types.Application {
	code, body, err := s.sendRequest("GET", "/v1/apps/"+id, nil)
	c.Assert(err, check.IsNil)
	c.Log(string(body))
	c.Assert(code, check.Equals, http.StatusOK)

	app := new(types.Application)
	err = s.bind(body, &app)
	c.Assert(err, check.IsNil)

	return app
}

func (s *ApiSuite) startApp(id string, c *check.C) {
	code, body, err := s.sendRequest("POST", "/v1/apps/"+id+"/start", nil)
	c.Assert(err, check.IsNil)
	c.Log(string(body))
	c.Assert(code, check.Equals, http.StatusAccepted)
}

func (s *ApiSuite) stopApp(id string, c *check.C) {
	code, body, err := s.sendRequest("POST", "/v1/apps/"+id+"/stop", nil)
	c.Assert(err, check.IsNil)
	c.Log(string(body))
	c.Assert(code, check.Equals, http.StatusAccepted)
}

func (s *ApiSuite) removeApp(id string, maxWait time.Duration, c *check.C) error {
	code, _, err := s.sendRequest("DELETE", "/v1/apps/"+id, nil)
	c.Assert(err, check.IsNil)
	c.Assert(code, check.Equals, http.StatusNoContent)

	timeout := time.After(maxWait)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return errWaitAppTimeout

		case <-ticker.C:
			if !s.existsApp(id, c) {
				return nil
			}
			c.Logf("waitting App %s to be removed ...", id)
		}
	}
}

func (s *ApiSuite) existsApp(id string, c *check.C) bool {
	code, body, err := s.sendRequest("GET", "/v1/apps/"+id, nil)
	c.Log(string(body))
	c.Assert(err, check.IsNil)

	matched, err := regexp.Match(".*node does not exist.*", body)
	return !(code == http.StatusNotFound && matched)
}

func (s *ApiSuite) scaleApp(id string, n int, c *check.C) {
	req := &types.Scale{
		Instances: n,
	}

	code, body, err := s.sendRequest("POST", "/v1/apps/"+id+"/scale", req)
	c.Assert(err, check.IsNil)
	c.Log(string(body))
	c.Assert(code, check.Equals, http.StatusAccepted)
}

// composes
func (s *ApiSuite) waitCompose(id, expect string, maxWait time.Duration, c *check.C) error {
	timeout := time.After(maxWait)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return errWaitComposeTimeout

		case <-ticker.C:
			cmp := s.inspectCompose(id, c)
			if cmp.OpStatus == expect {
				return nil
			}
			c.Logf("Compose: %s is %s ...", id, cmp.OpStatus)
		}
	}
}

func (s *ApiSuite) inspectCompose(id string, c *check.C) *types.ComposeWrapper {
	code, body, err := s.sendRequest("GET", "/v1/compose/"+id, nil)
	c.Assert(err, check.IsNil)
	c.Log(string(body))
	c.Assert(code, check.Equals, http.StatusOK)

	wrapper := new(types.ComposeWrapper)
	err = s.bind(body, &wrapper)
	c.Assert(err, check.IsNil)

	return wrapper
}

func (s *ApiSuite) removeCompose(id string, maxWait time.Duration, c *check.C) error {
	code, _, err := s.sendRequest("DELETE", "/v1/compose/"+id, nil)
	c.Assert(err, check.IsNil)
	c.Assert(code, check.Equals, http.StatusNoContent)

	timeout := time.After(maxWait)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return errWaitComposeTimeout

		case <-ticker.C:
			if !s.existsCompose(id, c) {
				return nil
			}
			c.Logf("waitting Compose %s to be removed ...", id)
		}
	}
}

func (s *ApiSuite) existsCompose(id string, c *check.C) bool {
	code, body, err := s.sendRequest("GET", "/v1/compose/"+id, nil)
	c.Log(string(body))
	c.Assert(err, check.IsNil)

	matched, err := regexp.Match(".*no such compose.*", body)
	return !(code == http.StatusNotFound && matched)
}

func (s *ApiSuite) runCompose(cmp *types.Compose, c *check.C) string {
	code, body, err := s.rawRunCompose(cmp)
	c.Assert(err, check.IsNil)
	c.Log(string(body))
	c.Assert(code, check.Equals, http.StatusCreated)

	var resp struct {
		Id string
	}
	err = s.bind(body, &resp)
	c.Assert(err, check.IsNil)
	c.Logf("created Compose%s", resp.Id)

	return resp.Id
}

func (s *ApiSuite) rawRunCompose(cmp *types.Compose) (int, []byte, error) {
	return s.sendRequest("POST", "/v1/compose", cmp)
}

// purge
//
func (s *ApiSuite) purge(maxWait time.Duration, c *check.C) error {
	code, _, err := s.sendRequest("POST", "/v1/purge", nil)
	c.Assert(err, check.IsNil)
	c.Assert(code, check.Equals, http.StatusNoContent)

	for goesby := int64(0); goesby <= int64(maxWait); goesby += int64(time.Second) {
		time.Sleep(time.Second)
		apps := s.listApps(c)
		if len(apps) == 0 {
			return nil
		}
	}

	return errWaitPurgeTimeout
}

// base
//
func (s *ApiSuite) sendRequest(method, uri string, data interface{}) (code int, body []byte, err error) {
	req, err := s.newRawReq(method, uri, data)
	if err != nil {
		return -1, nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return s.sendRawRequest(req)
}

func (s *ApiSuite) sendRawRequest(req *http.Request) (code int, body []byte, err error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, nil, err
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, nil, err
	}

	return resp.StatusCode, bs, nil
}

func (s *ApiSuite) newRawReq(method, uri string, data interface{}) (*http.Request, error) {
	buf := bytes.NewBuffer(nil)
	if data != nil {
		if err := json.NewEncoder(buf).Encode(data); err != nil {
			return nil, err
		}
	}

	return http.NewRequest(method, "http://"+s.SwanHost+uri, buf)
}

func (s *ApiSuite) bind(data []byte, val interface{}) error {
	return json.Unmarshal(data, &val)
}

// version builder
//
//
type verBuilder types.Version

func (b *verBuilder) Get() *types.Version {
	v := types.Version(*b)
	return &v
}

func (b *verBuilder) setCmd(cmd string) *verBuilder {
	b.Command = cmd
	return b
}

func (b *verBuilder) setName(name string) *verBuilder {
	b.Name = name
	return b
}

func (b *verBuilder) setCluster(cluster string) *verBuilder {
	b.Cluster = cluster
	return b
}

func (b *verBuilder) setRunAs(runas string) *verBuilder {
	b.RunAs = runas
	return b
}

func (b *verBuilder) setCount(n int) *verBuilder {
	b.Instances = int32(n)
	return b
}

func (b *verBuilder) setCPU(cpu float64) *verBuilder {
	b.CPUs = cpu
	return b
}

func (b *verBuilder) setGPU(gpu float64) *verBuilder {
	b.GPUs = gpu
	return b
}

func (b *verBuilder) setDisk(disk float64) *verBuilder {
	b.Disk = disk
	return b
}

func (b *verBuilder) setMem(mem float64) *verBuilder {
	b.Mem = mem
	return b
}

func (b *verBuilder) setImage(image string) *verBuilder {
	b.Container.Docker.Image = image
	return b
}

func (b *verBuilder) setPortMap(name, proto string, cp, hp int32) *verBuilder {
	pm := &types.PortMapping{
		Name:          name,
		Protocol:      proto,
		ContainerPort: cp,
		HostPort:      hp,
	}
	if b.Container.Docker.PortMappings == nil {
		b.Container.Docker.PortMappings = []*types.PortMapping{pm}
	} else {
		b.Container.Docker.PortMappings = append(b.Container.Docker.PortMappings, pm)
	}
	return b
}

func (b *verBuilder) setParameter(key, val string) *verBuilder {
	b.Container.Docker.Parameters = []*types.Parameter{
		{
			Key:   key,
			Value: val,
		},
	}
	return b
}

func (b *verBuilder) setConstraint(attr, op, val string) *verBuilder {
	b.Constraints = []*types.Constraint{
		{
			Attribute: attr,
			Operator:  op,
			Value:     val,
		},
	}
	return b
}

func (b *verBuilder) setContainer(typ string, docker *types.Docker) *verBuilder {
	b.Container.Type = typ
	b.Container.Docker = docker
	return b
}

func (b *verBuilder) setVolume(src, dst, mode string) *verBuilder {
	b.Container.Volumes = []*types.Volume{
		{
			HostPath:      src,
			ContainerPath: dst,
			Mode:          mode,
		},
	}
	return b
}

func (b *verBuilder) setHealthCheck(proto, pname, path, cmd string, c uint32, g, i, t, d float64) *verBuilder {
	b.HealthCheck = &types.HealthCheck{
		Protocol:            proto,
		PortName:            pname,
		Path:                path,
		Command:             cmd,
		ConsecutiveFailures: c,
		GracePeriodSeconds:  g,
		IntervalSeconds:     i,
		TimeoutSeconds:      t,
		DelaySeconds:        d,
	}
	return b
}

func (b *verBuilder) setNetwork(net string) *verBuilder {
	b.Container.Docker.Network = net
	return b
}

func (b *verBuilder) setIPs(ips []string) *verBuilder {
	b.IPs = ips
	return b
}

func demoVersion() *verBuilder {
	return &verBuilder{
		Name:        "demo",
		Instances:   int32(1),
		Command:     "",
		CPUs:        0.01,
		GPUs:        0,
		Mem:         5,
		Disk:        0,
		RunAs:       "integration",
		Cluster:     "",
		Constraints: nil,
		Container: &types.Container{
			Type: "docker",
			Docker: &types.Docker{
				Image:          "nginx",
				Network:        "bridge",
				Parameters:     nil,
				ForcePullImage: false,
				Privileged:     false,
				PortMappings: []*types.PortMapping{
					{
						Name:          "http",
						Protocol:      "tcp",
						ContainerPort: 80,
						HostPort:      80,
					},
				},
			},
		},
	}
}

// compose builder
//
//
type cmpBuilder types.Compose

func (b *cmpBuilder) Get() *types.Compose {
	v := types.Compose(*b)
	return &v
}

func (b *cmpBuilder) setName(name string) *cmpBuilder {
	b.Name = name
	return b
}

func (b *cmpBuilder) setDesc(desc string) *cmpBuilder {
	b.Desc = desc
	return b
}

func (b *cmpBuilder) setYAML(yaml string, extra map[string]*types.YamlExtra) *cmpBuilder {
	b.YAMLRaw = yaml
	b.YAMLExtra = extra
	return b
}

func demoCompose() *cmpBuilder {
	return &cmpBuilder{
		Name:      "demo",
		Desc:      "demo compose description",
		YAMLRaw:   "",
		YAMLEnv:   map[string]string{},
		YAMLExtra: map[string]*types.YamlExtra{},
	}
}

// yaml
//
var dockerYAMLHeader = `version: "3"

services:
`

var dockerServiceTemplate = ` {{.Name}}:
    image: {{.Image}}
    command: {{.Command}}
    network_mode: {{.NetworkMode}}

`

type DockerService struct {
	Name        string
	Image       string
	Command     string
	NetworkMode string
}

func demoYAML(count int, image, cmd, network string) (string, map[string]*types.YamlExtra) {
	buf := bytes.NewBufferString(dockerYAMLHeader)

	exts := make(map[string]*types.YamlExtra)
	for i := 1; i <= count; i++ {
		svrName := fmt.Sprintf("srv%d", i)
		ds := &DockerService{
			Name:        svrName,
			Image:       image,
			Command:     cmd,
			NetworkMode: network,
		}
		exts[svrName] = &types.YamlExtra{
			Resource: &types.Resource{0.01, 0, 10, 0, nil},
			RunAs:    "bbk",
			Labels:   map[string]string{"service_name": svrName},
		}
		tmpl, _ := template.New("").Parse(dockerServiceTemplate)
		tmpl.Execute(buf, ds)
	}

	return buf.String(), exts
}
