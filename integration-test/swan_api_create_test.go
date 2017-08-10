package main

import (
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
	err := s.purge(time.Second*30, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestCreateApp() purged")

	// New Create App
	//
	ver := demoVersion().setName("demo").setCount(10).setCPU(0.01).setMem(5).Get()
	id := s.createApp(ver, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*30, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestCreateApp() created")

	// verify app
	app := s.inspectApp(id, c)
	c.Assert(app.Name, check.Equals, "demo")
	c.Assert(app.TaskCount, check.Equals, 10)
	c.Assert(app.VersionCount, check.Equals, 1)
	c.Assert(len(app.Version), check.Equals, 1)

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
	err := s.purge(time.Second*30, c)
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
	var invalidNames = [][2]string{
		[2]string{"de..mo", "character.*not allowed"},
		[2]string{"<**$", "character.*not allowed"},
		[2]string{"   ", "character.*not allowed"},
		[2]string{strings.Repeat("x", 64), "appName empty or too long"},
		[2]string{"", "appName empty or too long"},
	}
	for _, pairs := range invalidNames {
		name, errmsg := pairs[0], pairs[1]
		ver := demoVersion().setName(name).Get()
		code, body, err := s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal app name verified")

	// invalid RunAs
	var invalidRunAses = [][2]string{
		[2]string{"bbk.", "character.*not allowed"},
		[2]string{"<**$", "character.*not allowed"},
		[2]string{"   ", "character.*not allowed"},
		[2]string{"", "runAs should not empty"},
	}
	for _, pairs := range invalidRunAses {
		runas, errmsg := pairs[0], pairs[1]
		ver := demoVersion().setRunAs(runas).Get()
		code, body, err := s.rawCreateApp(ver)
		c.Assert(err, check.IsNil)
		c.Assert(code, check.Equals, http.StatusBadRequest)
		c.Log(string(body))
		match, _ = regexp.MatchString(errmsg, string(body))
		c.Assert(match, check.Equals, true)
	}
	fmt.Println("TestCreateInvalidApp() illegal app runAs verified")

	// invalid image
	ver := demoVersion().setImage("").Get()
	code, body, err = s.rawCreateApp(ver)
	c.Assert(err, check.IsNil)
	c.Assert(code, check.Equals, http.StatusBadRequest)
	c.Log(string(body))
	match, _ = regexp.MatchString("image field required", string(body))
	c.Assert(match, check.Equals, true)
	fmt.Println("TestCreateInvalidApp() illegal app image verified")

	// invalid instances
	ver = demoVersion().setCount(0).Get()
	code, body, err = s.rawCreateApp(ver)
	c.Assert(err, check.IsNil)
	c.Assert(code, check.Equals, http.StatusBadRequest)
	c.Log(string(body))
	match, _ = regexp.MatchString("should greater than 0", string(body))
	c.Assert(match, check.Equals, true)
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
}
