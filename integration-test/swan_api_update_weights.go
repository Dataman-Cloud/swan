package main

import (
	"time"

	check "gopkg.in/check.v1"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *ApiSuite) TestUpdateWeights(c *check.C) {
	// purge
	startAt := time.Now()
	err := s.purge(time.Second*60, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestUpdateWeights() purged", startAt)

	// create app
	startAt = time.Now()
	ver := demoVersion().setName("demo").setCount(5).setCPU(0.01).setMem(5).setProxy(true, "www.xxx.com", "", false).Get()
	id := s.createApp(ver, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*180, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestUpdateWeights() created", startAt)

	// verify app
	app := s.inspectApp(id, c)
	c.Assert(app.Name, check.Equals, "demo")
	c.Assert(app.TaskCount, check.Equals, 5)
	c.Assert(app.VersionCount, check.Equals, 1)
	c.Assert(len(app.Version), check.Equals, 1)
	c.Assert(app.ErrMsg, check.Equals, "")

	// verify proxy record
	proxy := s.listAppProxies(id, c)
	c.Assert(proxy, check.Not(check.IsNil))
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

	// do canay update first
	body := &types.CanaryUpdateBody{
		Version:   demoVersion().setName("demo").setCount(5).setCPU(0.01).setMem(10).setProxy(true, "www.xxx.com", "", false).Get(),
		Instances: 3,
		Value:     0.5,
		OnFailure: "continue",
		Delay:     0.5,
	}
	s.canaryUpdate(id, body, c)
	err = s.waitApp(id, types.OpStatusCanaryUnfinished, time.Second*180, c)
	c.Assert(err, check.IsNil)

	// verify app
	app = s.inspectApp(id, c)
	c.Assert(app.Name, check.Equals, "demo")
	c.Assert(app.VersionCount, check.Equals, 2)
	c.Assert(len(app.Version), check.Equals, 2)
	c.Assert(app.OpStatus, check.Equals, types.OpStatusCanaryUnfinished)
	c.Assert(app.ErrMsg, check.Equals, "")

	// verify app tasks
	tasks := s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 5)

	var n, m int
	for _, task := range tasks {
		if task.Weight == 67 {
			n++
		}

		if task.Weight == 100 {
			m++
		}
	}

	c.Assert(n, check.Equals, 3)
	c.Assert(m, check.Equals, 2)

	// verify app versions
	vers := s.listAppVersions(id, c)
	c.Assert(len(vers), check.Equals, 2)
	c.Assert(vers[0].Mem, check.Equals, float64(10))

	counter := make(map[string]int)
	for _, task := range tasks {
		if v, ok := counter[task.Version]; ok {
			v++
			counter[task.Version] = v
		} else {
			counter[task.Version] = 1
		}
	}

	c.Assert(counter[vers[0].ID], check.Equals, 3)
	c.Assert(counter[vers[1].ID], check.Equals, 2)

	// verify proxy record
	proxy = s.listAppProxies(id, c)
	c.Assert(proxy, check.Not(check.IsNil))
	c.Assert(proxy.Alias, check.Equals, "www.xxx.com")
	c.Assert(len(proxy.Backends), check.Equals, 5)
	c.Assert(proxy.Listen, check.Equals, "")
	c.Assert(proxy.Sticky, check.Equals, false)
	var y int
	for _, b := range proxy.Backends {
		if b.Weight == 67 {
			y++
		}
	}
	c.Assert(y, check.Equals, 3)

	// verify dns records
	dns = s.listAppDNS(id, c)
	c.Assert(len(dns), check.Equals, 5)

	var x int
	for _, d := range dns {
		if d.Weight == 67 {
			x++
		}
		c.Assert(d.Port, check.Not(check.Equals), "")
	}
	c.Assert(x, check.Equals, 3)

	// Update weights again
	startAt = time.Now()
	s.updateWeights(id, &types.UpdateWeightsBody{
		Value: 0.8,
	}, c)
	err = s.waitApp(id, types.OpStatusCanaryUnfinished, time.Second*180, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestUpdateWeights() updated", startAt)

	// verify app
	app = s.inspectApp(id, c)
	c.Assert(app.Name, check.Equals, "demo")
	c.Assert(app.VersionCount, check.Equals, 2)
	c.Assert(len(app.Version), check.Equals, 2)
	c.Assert(app.OpStatus, check.Equals, types.OpStatusCanaryUnfinished)
	c.Assert(app.ErrMsg, check.Equals, "")

	// verify app tasks
	tasks = s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 5)

	new := 0
	for _, task := range tasks {
		if task.Version == vers[0].ID {
			if task.Weight == 267 {
				new++
			}
		}
	}

	c.Assert(new, check.Equals, 3)

	// verify proxy record
	proxy = s.listAppProxies(id, c)
	c.Assert(proxy, check.Not(check.IsNil))
	c.Assert(proxy.Alias, check.Equals, "www.xxx.com")
	c.Assert(len(proxy.Backends), check.Equals, 5)
	c.Assert(proxy.Listen, check.Equals, "")
	c.Assert(proxy.Sticky, check.Equals, false)
	var y1 int
	for _, b := range proxy.Backends {
		if b.Weight == 267 {
			y1++
		}
	}
	c.Assert(y1, check.Equals, 3)

	// verify dns records
	dns = s.listAppDNS(id, c)
	c.Assert(len(dns), check.Equals, 5)

	var x1 int
	for _, d := range dns {
		if d.Weight == 267 {
			x1++
		}
		c.Assert(d.Port, check.Not(check.Equals), "")
	}
	c.Assert(x1, check.Equals, 3)

	// switch all traffics
	startAt = time.Now()
	s.updateWeights(id, &types.UpdateWeightsBody{
		Value: 1.0,
	}, c)
	err = s.waitApp(id, types.OpStatusCanaryUnfinished, time.Second*180, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestUpdateWeights() updated", startAt)

	// verify app
	app = s.inspectApp(id, c)
	c.Assert(app.Name, check.Equals, "demo")
	c.Assert(app.VersionCount, check.Equals, 2)
	c.Assert(len(app.Version), check.Equals, 2)
	c.Assert(app.OpStatus, check.Equals, types.OpStatusCanaryUnfinished)
	c.Assert(app.ErrMsg, check.Equals, "")

	// verify app tasks
	tasks = s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 5)

	new, old := 0, 0
	for _, task := range tasks {
		if task.Version == vers[0].ID {
			if task.Weight == 100 {
				new++
			}
		}

		if task.Weight == 0 {
			old++
		}
	}

	c.Assert(new, check.Equals, 3)
	c.Assert(old, check.Equals, 2)

	// verify proxy record
	proxy = s.listAppProxies(id, c)
	c.Assert(proxy, check.Not(check.IsNil))
	c.Assert(proxy.Alias, check.Equals, "www.xxx.com")
	c.Assert(len(proxy.Backends), check.Equals, 5)
	c.Assert(proxy.Listen, check.Equals, "")
	c.Assert(proxy.Sticky, check.Equals, false)
	var m1, n1 int
	for _, b := range proxy.Backends {
		if b.Weight == 100 {
			m1++
		}

		if b.Weight == 0 {
			n1++
		}
	}
	c.Assert(m1, check.Equals, 3)
	c.Assert(n1, check.Equals, 2)

	// verify dns records
	dns = s.listAppDNS(id, c)
	c.Assert(len(dns), check.Equals, 5)

	var x2, y2 int
	for _, d := range dns {
		if d.Weight == 100 {
			x2++
		}

		if d.Weight == 0 {
			y2++
		}

		c.Assert(d.Port, check.Not(check.Equals), "")
	}
	c.Assert(x2, check.Equals, 3)
	c.Assert(y2, check.Equals, 2)

	// clean up
	startAt = time.Now()
	err = s.removeApp(id, time.Second*10, c)
	c.Assert(err, check.IsNil)
	costPrintln("TestCanaryUpdate() removed", startAt)
}
