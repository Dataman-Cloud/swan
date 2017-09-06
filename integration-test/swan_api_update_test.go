package main

import (
	"time"

	check "gopkg.in/check.v1"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *ApiSuite) TestUpdateApp(c *check.C) {
	// Purge
	//
	startAt := time.Now()
	err := s.purge(time.Second*60, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestUpdateApp() purged", startAt)

	// New Create App
	//
	startAt = time.Now()
	ver := demoVersion().setName("demo").setCount(3).setCPU(0.01).setMem(5).setProxy(true, "www.xxx.com", "", false).Get()
	id := s.createApp(ver, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestUpdateApp() created", startAt)

	// verify app
	app := s.inspectApp(id, c)
	c.Assert(app.Name, check.Equals, "demo")
	c.Assert(app.TaskCount, check.Equals, 3)
	c.Assert(app.VersionCount, check.Equals, 1)
	c.Assert(len(app.Version), check.Equals, 1)
	c.Assert(app.ErrMsg, check.Equals, "")

	// verify app versions
	vers := s.listAppVersions(id, c)
	c.Assert(len(vers), check.Equals, 1)
	c.Assert(vers[0].CPUs, check.Equals, 0.01)
	c.Assert(vers[0].Mem, check.Equals, float64(5))
	c.Assert(vers[0].Instances, check.Equals, int32(3))
	c.Assert(vers[0].RunAs, check.Equals, app.RunAs)

	// verify app tasks
	tasks := s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 3)
	for _, task := range tasks {
		c.Assert(task.Version, check.Equals, vers[0].ID)
	}

	// verify proxy record
	proxy := s.listAppProxies(id, c)
	c.Assert(proxy, check.Not(check.IsNil))
	c.Assert(proxy.Alias, check.Equals, "www.xxx.com")
	c.Assert(len(proxy.Backends), check.Equals, 3)
	c.Assert(proxy.Listen, check.Equals, "")
	c.Assert(proxy.Sticky, check.Equals, false)
	for _, b := range proxy.Backends {
		c.Assert(b.Weight, check.Equals, float64(100))
	}

	// verify dns records
	dns := s.listAppDNS(id, c)
	c.Assert(len(dns), check.Equals, 3)
	for _, d := range dns {
		c.Assert(d.Weight, check.Equals, float64(100))
		c.Assert(d.Port, check.Not(check.Equals), "")
	}

	// Update App
	//
	startAt = time.Now()
	newVer := demoVersion().setName("demo").setCount(3).setCPU(0.02).setMem(10).setProxy(true, "www.xxx.com", "", false).Get()
	s.updateApp(id, newVer, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestUpdateApp() updated", startAt)

	// verify app
	app = s.inspectApp(id, c)
	c.Assert(app.Name, check.Equals, "demo")
	c.Assert(app.TaskCount, check.Equals, 3)
	c.Assert(app.VersionCount, check.Equals, 2)
	c.Assert(len(app.Version), check.Equals, 1)
	c.Assert(app.ErrMsg, check.Equals, "")

	// verify app versions
	vers = s.listAppVersions(id, c)
	c.Assert(len(vers), check.Equals, 2)
	c.Assert(vers[0].CPUs, check.Equals, 0.02)
	c.Assert(vers[0].Mem, check.Equals, float64(10))
	c.Assert(vers[0].Instances, check.Equals, int32(3))
	c.Assert(vers[0].RunAs, check.Equals, app.RunAs)

	// verify app tasks
	tasks = s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 3)
	for _, task := range tasks {
		c.Assert(task.Version, check.Equals, vers[0].ID)
	}

	// verify proxy record
	proxy = s.listAppProxies(id, c)
	c.Assert(proxy, check.Not(check.IsNil))
	c.Assert(proxy.Alias, check.Equals, "www.xxx.com")
	c.Assert(len(proxy.Backends), check.Equals, 3)
	c.Assert(proxy.Listen, check.Equals, "")
	c.Assert(proxy.Sticky, check.Equals, false)
	for _, b := range proxy.Backends {
		c.Assert(b.Weight, check.Equals, float64(100))
	}

	// verify dns records
	dns = s.listAppDNS(id, c)
	c.Assert(len(dns), check.Equals, 3)
	for _, d := range dns {
		c.Assert(d.Weight, check.Equals, float64(100))
		c.Assert(d.Port, check.Not(check.Equals), "")
	}

	// Update App Again
	//
	startAt = time.Now()
	newVer = demoVersion().setName("demo").setCount(3).setCPU(0.03).setMem(7).setProxy(true, "www.xxx.com", "", false).Get()
	s.updateApp(id, newVer, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestUpdateApp() updated", startAt)

	// verify app
	app = s.inspectApp(id, c)
	c.Assert(app.Name, check.Equals, "demo")
	c.Assert(app.TaskCount, check.Equals, 3)
	c.Assert(app.VersionCount, check.Equals, 3)
	c.Assert(len(app.Version), check.Equals, 1)
	c.Assert(app.ErrMsg, check.Equals, "")

	// verify app versions
	vers = s.listAppVersions(id, c)
	c.Assert(len(vers), check.Equals, 3)
	c.Assert(vers[0].CPUs, check.Equals, 0.03)
	c.Assert(vers[0].Mem, check.Equals, float64(7))
	c.Assert(vers[0].Instances, check.Equals, int32(3))
	c.Assert(vers[0].RunAs, check.Equals, app.RunAs)

	// verify app tasks
	tasks = s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 3)
	for _, task := range tasks {
		c.Assert(task.Version, check.Equals, vers[0].ID)
	}

	// verify proxy record
	proxy = s.listAppProxies(id, c)
	c.Assert(proxy, check.Not(check.IsNil))
	c.Assert(proxy.Alias, check.Equals, "www.xxx.com")
	c.Assert(len(proxy.Backends), check.Equals, 3)
	c.Assert(proxy.Listen, check.Equals, "")
	c.Assert(proxy.Sticky, check.Equals, false)
	for _, b := range proxy.Backends {
		c.Assert(b.Weight, check.Equals, float64(100))
	}

	// verify dns records
	dns = s.listAppDNS(id, c)
	c.Assert(len(dns), check.Equals, 3)
	for _, d := range dns {
		c.Assert(d.Weight, check.Equals, float64(100))
		c.Assert(d.Port, check.Not(check.Equals), "")
	}

	// Remove
	//
	startAt = time.Now()
	err = s.removeApp(id, time.Second*10, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestUpdateApp() removed", startAt)
}
