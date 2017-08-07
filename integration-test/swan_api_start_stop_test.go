package main

import (
	"fmt"
	"time"

	check "gopkg.in/check.v1"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *ApiSuite) TestStartStopApp(c *check.C) {
	// Purge
	//
	err := s.purge(time.Second*30, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestStartStopApp() purged")

	// New Create App
	//
	ver := demoVersion().setName("demo").setCount(10).setCPU(0.01).setMem(5).Get()
	id := s.createApp(ver, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*30, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestStartStopApp() created")

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

	// Stop App
	//
	s.stopApp(id, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*10, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestStartStopApp() stopped")

	app = s.inspectApp(id, c)
	c.Assert(app.VersionCount, check.Equals, 1)

	tasks = s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 0)

	// Start App
	//
	s.startApp(id, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*10, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestStartStopApp() started")

	app = s.inspectApp(id, c)
	c.Assert(app.VersionCount, check.Equals, 1)

	tasks = s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 10)

	// Stop App Again
	//
	s.stopApp(id, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*10, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestStartStopApp() stopped")

	app = s.inspectApp(id, c)
	c.Assert(app.VersionCount, check.Equals, 1)

	tasks = s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 0)

	// Start App Again
	//
	s.startApp(id, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*10, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestStartStopApp() started")

	app = s.inspectApp(id, c)
	c.Assert(app.VersionCount, check.Equals, 1)

	tasks = s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 10)

	// Remove
	//
	err = s.removeApp(id, time.Second*10, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestStartStopApp() removed")
}
