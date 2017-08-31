package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	check "gopkg.in/check.v1"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *ApiSuite) TestCreateApp(c *check.C) {
	// Purge
	//
	err := s.purge(time.Second*60, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestCreateApp() purged")

	// New Create App
	//
	ver := demoVersion().setName("demo").setCount(10).setCPU(0.01).setMem(5).Get()
	id := s.createApp(ver, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestCreateApp() created")

	// verify app
	app := s.inspectApp(id, c)
	c.Assert(app.Name, check.Equals, "demo")
	c.Assert(app.TaskCount, check.Equals, 10)
	c.Assert(app.VersionCount, check.Equals, 1)
	c.Assert(len(app.Version), check.Equals, 1)
	c.Assert(app.ErrMsg, check.Equals, "")

	// verify app versions
	vers := s.listAppVersions(id, c)
	c.Assert(len(vers), check.Equals, 1)
	c.Assert(vers[0].CPUs, check.Equals, 0.01)
	c.Assert(vers[0].Mem, check.Equals, float64(5))
	c.Assert(vers[0].Instances, check.Equals, int32(10))
	c.Assert(vers[0].RunAs, check.Equals, app.RunAs)

	// verify app tasks
	tasks := s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 10)
	for _, task := range tasks {
		c.Assert(task.Version, check.Equals, vers[0].ID)
	}

	// Remove
	//
	err = s.removeApp(id, time.Second*10, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestCreateApp() removed")
}

func (s *ApiSuite) TestCreateInvalidApp(c *check.C) {
	// Purge
	//
	err := s.purge(time.Second*60, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestCreateInvalidApp() purged")

	// Invalid Create Request
	//

	// invalid Content-Type
	req, err := s.newRawReq("POST", "/v1/apps", "....")
	c.Assert(err, check.IsNil)
	code, body, err := s.sendRawRequest(req)
	c.Assert(err, check.IsNil)
	c.Assert(code, check.Equals, http.StatusBadRequest)
	c.Log(string(body))
	match, _ := regexp.MatchString("must be.*application/json", string(body))
	c.Assert(match, check.Equals, true)
	fmt.Println("TestCreateInvalidApp() illegal content type verified")

	// invalid App Name
	var invalidNames = map[*types.Version]string{
		demoVersion().setName("de..mo").Get():                "character.*not allowed",
		demoVersion().setName("<**$").Get():                  "character.*not allowed",
		demoVersion().setName("    ").Get():                  "character.*not allowed",
		demoVersion().setName(strings.Repeat("x", 65)).Get(): "appName should between",
		demoVersion().setName("").Get():                      "appName should between",
	}
	for ver, errmsg := range invalidNames {
		code, body, err := s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal app name verified")

	// invalid RunAs
	var invalidRunAses = map[*types.Version]string{
		demoVersion().setRunAs("de..mo").Get():                "character.*not allowed",
		demoVersion().setRunAs("<**$").Get():                  "character.*not allowed",
		demoVersion().setRunAs("    ").Get():                  "character.*not allowed",
		demoVersion().setRunAs(strings.Repeat("x", 65)).Get(): "runAs length should between",
		demoVersion().setRunAs("").Get():                      "runAs length should between",
	}
	for ver, errmsg := range invalidRunAses {
		code, body, err := s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal app runAs verified")

	// invalid Cluster
	var invalidClusters = map[*types.Version]string{
		demoVersion().setCluster("de..mo").Get():                "character.*not allowed",
		demoVersion().setCluster("<**$").Get():                  "character.*not allowed",
		demoVersion().setCluster("    ").Get():                  "character.*not allowed",
		demoVersion().setCluster(strings.Repeat("x", 65)).Get(): "cluster name length should between",
	}
	for ver, errmsg := range invalidClusters {
		code, body, err := s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal cluster verified")

	// invalid image
	ver := demoVersion().setImage("").Get()
	code, body, err = s.rawCreateApp(ver)
	c.Assert(err, check.IsNil)
	c.Assert(code, check.Equals, http.StatusBadRequest)
	c.Log(string(body))
	match, _ = regexp.MatchString("image required", string(body))
	c.Assert(match, check.Equals, true)
	fmt.Println("TestCreateInvalidApp() illegal app image verified")

	// invalid network
	ver = demoVersion().setNetwork("").Get()
	code, body, err = s.rawCreateApp(ver)
	c.Assert(err, check.IsNil)
	c.Assert(code, check.Equals, http.StatusBadRequest)
	c.Log(string(body))
	match, _ = regexp.MatchString("network required", string(body))
	c.Assert(match, check.Equals, true)
	fmt.Println("TestCreateInvalidApp() illegal network verified")

	// invalid instances
	var invalidCounts = map[*types.Version]string{
		demoVersion().setCount(0).Get():  "instance count must be positive",
		demoVersion().setCount(-1).Get(): "instance count must be positive",
	}
	for ver, errmsg := range invalidCounts {
		code, body, err = s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal app count verified")

	// invalid resource values
	var invalidResources = map[*types.Version]string{
		demoVersion().setCPU(-1).Get():  "cpus can't be negative",
		demoVersion().setGPU(-1).Get():  "gpus can't be negative",
		demoVersion().setMem(-1).Get():  "memory can't be negative",
		demoVersion().setDisk(-1).Get(): "disk can't be negative",
	}
	for ver, errmsg := range invalidResources {
		code, body, err = s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal resources verified")

	// invalid port mapping
	var invalidPortMaps = map[*types.Version]string{
		demoVersion().setPortMap("web", "tcp", 10, 10).setPortMap("web", "tcp", 11, 11).Get(): "port name .* conflict",
		demoVersion().setPortMap("", "tcp", 10, 10).Get():                                     "port must be named",
		demoVersion().setPortMap("web", "xxx", 10, 10).Get():                                  "unsupported port protocol",
		demoVersion().setPortMap("web", "udp", -1, 10).Get():                                  "container port out of range",
		demoVersion().setPortMap("web", "udp", 65536, 10).Get():                               "container port out of range",
		demoVersion().setPortMap("web", "udp", 10, -1).Get():                                  "host port out of range",
		demoVersion().setPortMap("web", "udp", 10, 65536).Get():                               "host port out of range",
	}
	for ver, errmsg := range invalidPortMaps {
		code, body, err = s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal portmapping verified")

	// invalid parameters
	var invalidParameters = map[*types.Version]string{
		demoVersion().setParameter("", "host").Get(): "Parameter.Key required",
	}
	for ver, errmsg := range invalidParameters {
		code, body, err = s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal parameter verified")

	// invalid constraint
	var invalidConstraints = map[*types.Version]string{
		demoVersion().setConstraint("", "==", "").Get():     "attribute required for constraint",
		demoVersion().setConstraint("xxx", "xxx", "").Get(): "Operator not supported",
	}
	for ver, errmsg := range invalidConstraints {
		code, body, err = s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal constraint verified")

	// invalid ips
	var invalidIPs = map[*types.Version]string{
		demoVersion().setNetwork("swan").setCount(1).setIPs([]string{"xx"}).Get():                         "invalid ip:",
		demoVersion().setNetwork("swan").setCount(1).setIPs([]string{"127.0.0.1"}).Get():                  "invalid ip:",
		demoVersion().setNetwork("swan").setCount(2).setIPs([]string{"192.168.1.1", "192.168.1.1"}).Get(): "ip.*conflict",
	}

	for ver, errmsg := range invalidIPs {
		code, body, err = s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal ips verified")

	// invalid proxy
	var invalidProxies = map[string]string{
		`{"enabled":true, "listen": -1}`:    "proxy.Listen out of range",
		`{"enabled":true, "listen": 65536}`: "proxy.Listen out of range",
	}
	for bs, errmsg := range invalidProxies {
		err := json.Unmarshal([]byte(bs), new(types.Proxy))
		c.Assert(err, check.Not(check.IsNil))
		match, _ = regexp.MatchString(errmsg, err.Error())
	}
	fmt.Println("TestCreateInvalidApp() illegal proxy verified")

	// invalid healthcheck
	var invalidHealthChecks = map[*types.Version]string{

		demoVersion().setHealthCheck("xxx", "web", "/", "xx", 10, 1, 1, 1, 1).Get():                                                            "unsupported health check protocol",
		demoVersion().setHealthCheck("", "web", "/", "xx", 10, 1, 1, 1, 1).Get():                                                               "unsupported health check protocol",
		demoVersion().setHealthCheck("cmd", "", "", "", 10, 1, 1, 1, 1).Get():                                                                  "command required for cmd health check",
		demoVersion().setHealthCheck("tcp", "", "", "", 10, 1, 1, 1, 1).Get():                                                                  "port name required for tcp health check",
		demoVersion().setHealthCheck("http", "", "", "", 10, 1, 1, 1, 1).Get():                                                                 "port name required for http health check",
		demoVersion().setHealthCheck("http", "web", "", "", 10, 1, 1, 1, 1).Get():                                                              "path required for http health check",
		demoVersion().setHealthCheck("http", "web", "/", "xx", 10, -1, 1, 1, 1).Get():                                                          "gracePeriodSeconds can't be negative",
		demoVersion().setHealthCheck("http", "web", "/", "xx", 10, 1, 0, 1, 1).Get():                                                           "intervalSeconds should be positive",
		demoVersion().setHealthCheck("http", "web", "/", "xx", 10, 1, -1, 1, 1).Get():                                                          "intervalSeconds should be positive",
		demoVersion().setHealthCheck("http", "web", "/", "xx", 10, 1, 1, -1, 1).Get():                                                          "timeoutSeconds can't be negative",
		demoVersion().setHealthCheck("http", "web", "/", "xx", 10, 1, 1, 1, -1).Get():                                                          "delaySeconds can't be negative",
		demoVersion().setHealthCheck("http", "web", "/", "xx", 10, 1, -1, 1, 1).Get():                                                          "intervalSeconds should be positive",
		demoVersion().setHealthCheck("cmd", "", "", "ps", 10, 1, 1, 1, 1).setNetwork("swan").setCount(1).setIPs([]string{"192.168.1.1"}).Get(): "can't use cmd health check on fixed type app",
	}
	for ver, errmsg := range invalidHealthChecks {
		code, body, err := s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal health check verified")

	// invalid container type
	var invalidContainerTypes = map[*types.Version]string{
		demoVersion().setContainer("lxc", nil).Get():    "only support docker containerization",
		demoVersion().setContainer("mesos", nil).Get():  "only support docker containerization",
		demoVersion().setContainer("docker", nil).Get(): "docker containerization settings required",
	}
	for ver, errmsg := range invalidContainerTypes {
		code, body, err := s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal container type verified")

	// invalid volumes
	var invalidVolumes = map[*types.Version]string{
		demoVersion().setVolume("..", "/data", "RO").Get(): "Volume.HostPath should be absolute path",
		demoVersion().setVolume("/", "/data", "xx").Get():  "unsupported Volume.Mode",
		demoVersion().setVolume("/", "/data", "ro").Get():  "unsupported Volume.Mode",
		demoVersion().setVolume("/", "/data", "rw").Get():  "unsupported Volume.Mode",
	}
	for ver, errmsg := range invalidVolumes {
		code, body, err := s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal volumes verified")
}
