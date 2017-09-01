package main

import (
	"time"

	check "gopkg.in/check.v1"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *ApiSuite) TestStartStopApp(c *check.C) {
	// Purge
	//
	startAt := time.Now()
	err := s.purge(time.Second*60, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestStartStopApp() purged", startAt)

	// New Create App
	//
	startAt = time.Now()
	ver := demoVersion().setName("demo").setCount(10).setCPU(0.01).setMem(5).setProxy(true, "www.xxx.com", "", false).Get()
	id := s.createApp(ver, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestStartStopApp() created", startAt)

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

	// verify proxy record
	proxy := s.listAppProxies(id, c)
	c.Assert(proxy.Alias, check.Equals, "www.xxx.com")
	c.Assert(len(proxy.Backends), check.Equals, 10)
	c.Assert(proxy.Listen, check.Equals, "")
	c.Assert(proxy.Sticky, check.Equals, false)
	for _, b := range proxy.Backends {
		c.Assert(b.Weight, check.Equals, float64(100))
	}

	// verify dns records
	dns := s.listAppDNS(id, c)
	c.Assert(len(dns), check.Equals, 10)
	for _, d := range dns {
		c.Assert(d.Weight, check.Equals, float64(100))
		c.Assert(d.Port, check.Not(check.Equals), "")
	}

	// Stop App
	//
	startAt = time.Now()
	s.stopApp(id, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestStartStopApp() stopped", startAt)

	app = s.inspectApp(id, c)
	c.Assert(app.VersionCount, check.Equals, 1)
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

	// Start App
	//
	startAt = time.Now()
	s.startApp(id, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestStartStopApp() started", startAt)

	app = s.inspectApp(id, c)
	c.Assert(app.VersionCount, check.Equals, 1)
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

	// Stop App Again
	//
	startAt = time.Now()
	s.stopApp(id, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestStartStopApp() stopped", startAt)

	app = s.inspectApp(id, c)
	c.Assert(app.VersionCount, check.Equals, 1)
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

	// Start App Again
	//
	startAt = time.Now()
	s.startApp(id, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestStartStopApp() started", startAt)

	app = s.inspectApp(id, c)
	c.Assert(app.VersionCount, check.Equals, 1)
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

	// Remove
	//
	startAt = time.Now()
	err = s.removeApp(id, time.Second*10, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestStartStopApp() removed", startAt)
}
