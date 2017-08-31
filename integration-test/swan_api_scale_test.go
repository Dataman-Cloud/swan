package main

import (
	"fmt"
	"time"

	check "gopkg.in/check.v1"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *ApiSuite) TestScaleApp(c *check.C) {
	// Purge
	//
	err := s.purge(time.Second*60, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestScaleApp() purged")

	// New Create App
	//
	ver := demoVersion().setName("demo").setCount(5).setCPU(0.01).setMem(5).setProxy(true, "www.xxx.com", "", false).Get()
	id := s.createApp(ver, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestScaleApp() created")

	// verify app
	app := s.inspectApp(id, c)
	c.Assert(app.Name, check.Equals, "demo")
	c.Assert(app.TaskCount, check.Equals, 5)
	c.Assert(app.VersionCount, check.Equals, 1)
	c.Assert(len(app.Version), check.Equals, 1)
	c.Assert(app.ErrMsg, check.Equals, "")

	// verify app versions
	vers := s.listAppVersions(id, c)
	c.Assert(len(vers), check.Equals, 1)
	c.Assert(vers[0].CPUs, check.Equals, 0.01)
	c.Assert(vers[0].Mem, check.Equals, float64(5))
	c.Assert(vers[0].Instances, check.Equals, int32(5))
	c.Assert(vers[0].RunAs, check.Equals, app.RunAs)

	// verify app tasks
	tasks := s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 5)
	for _, task := range tasks {
		c.Assert(task.Version, check.Equals, vers[0].ID)
	}

	// verify proxy record
	proxy := s.listAppProxies(id, c)
	c.Assert(proxy.Alias, check.Equals, "www.xxx.com")
	c.Assert(len(proxy.Backends), check.Equals, 5)
	c.Assert(proxy.Listen, check.Equals, "")
	c.Assert(proxy.Sticky, check.Equals, false)
	for _, b := range proxy.Backends {
		c.Assert(b.Weight, check.Equals, float64(100))
	}

	// verify dns records
	dns := s.listAppDNS(id, c)
	c.Assert(len(dns), check.Equals, 5)
	for _, d := range dns {
		c.Assert(d.Weight, check.Equals, float64(100))
		c.Assert(d.Port, check.Not(check.Equals), "")
	}

	// Scale Up App
	//
	s.scaleApp(id, 10, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestScaleApp() scaled up")

	app = s.inspectApp(id, c)
	c.Assert(app.VersionCount, check.Equals, 2)
	c.Assert(app.ErrMsg, check.Equals, "")

	tasks = s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 10)

	// verify proxy record
	proxy = s.listAppProxies(id, c)
	c.Assert(proxy.Alias, check.Equals, "www.xxx.com")
	c.Assert(len(proxy.Backends), check.Equals, 10)
	c.Assert(proxy.Listen, check.Equals, "")
	c.Assert(proxy.Sticky, check.Equals, false)
	for _, b := range proxy.Backends {
		c.Assert(b.Weight, check.Equals, float64(100))
	}

	// verify dns records
	dns = s.listAppDNS(id, c)
	c.Assert(len(dns), check.Equals, 10)
	for _, d := range dns {
		c.Assert(d.Weight, check.Equals, float64(100))
		c.Assert(d.Port, check.Not(check.Equals), "")
	}

	// Scale Down App
	//
	s.scaleApp(id, 1, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestScaleApp() scaled down")

	app = s.inspectApp(id, c)
	c.Assert(app.VersionCount, check.Equals, 3)
	c.Assert(app.ErrMsg, check.Equals, "")

	tasks = s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 1)

	// verify proxy record
	proxy = s.listAppProxies(id, c)
	c.Assert(proxy.Alias, check.Equals, "www.xxx.com")
	c.Assert(len(proxy.Backends), check.Equals, 1)
	c.Assert(proxy.Listen, check.Equals, "")
	c.Assert(proxy.Sticky, check.Equals, false)
	for _, b := range proxy.Backends {
		c.Assert(b.Weight, check.Equals, float64(100))
	}

	// verify dns records
	dns = s.listAppDNS(id, c)
	c.Assert(len(dns), check.Equals, 1)
	for _, d := range dns {
		c.Assert(d.Weight, check.Equals, float64(100))
		c.Assert(d.Port, check.Not(check.Equals), "")
	}

	// Scale Up App Again
	//
	s.scaleApp(id, 5, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestScaleApp() scaled up")

	app = s.inspectApp(id, c)
	c.Assert(app.VersionCount, check.Equals, 4)
	c.Assert(app.ErrMsg, check.Equals, "")

	tasks = s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 5)

	// verify proxy record
	proxy = s.listAppProxies(id, c)
	c.Assert(proxy.Alias, check.Equals, "www.xxx.com")
	c.Assert(len(proxy.Backends), check.Equals, 5)
	c.Assert(proxy.Listen, check.Equals, "")
	c.Assert(proxy.Sticky, check.Equals, false)
	for _, b := range proxy.Backends {
		c.Assert(b.Weight, check.Equals, float64(100))
	}

	// verify dns records
	dns = s.listAppDNS(id, c)
	c.Assert(len(dns), check.Equals, 5)
	for _, d := range dns {
		c.Assert(d.Weight, check.Equals, float64(100))
		c.Assert(d.Port, check.Not(check.Equals), "")
	}

	// Scale Down App Again
	//
	s.scaleApp(id, 0, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestScaleApp() scaled down")

	app = s.inspectApp(id, c)
	c.Assert(app.VersionCount, check.Equals, 5)
	c.Assert(app.ErrMsg, check.Equals, "")

	tasks = s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 0)

	// verify proxy record
	proxy = s.listAppProxies(id, c)
	c.Assert(proxy.Alias, check.Equals, "")
	c.Assert(len(proxy.Backends), check.Equals, 0)
	c.Assert(proxy.Listen, check.Equals, "")
	c.Assert(proxy.Sticky, check.Equals, false)

	// verify dns records
	dns = s.listAppDNS(id, c)
	c.Assert(len(dns), check.Equals, 0)

	// Remove
	//
	err = s.removeApp(id, time.Second*10, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestScaleApp() removed")
}
