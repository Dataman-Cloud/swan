package main

import (
	"fmt"
	"time"

	check "gopkg.in/check.v1"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *ApiSuite) TestRollBackApp(c *check.C) {
	// Purge
	//
	err := s.purge(time.Second*30, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestUpdateApp() purged")

	// New Create App
	//
	ver := demoVersion().setName("demo").setCount(3).setCPU(0.01).setMem(5).Get()
	id := s.createApp(ver, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*30, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestUpdateApp() created")

	// verify app
	app := s.inspectApp(id, c)
	c.Assert(app.Name, check.Equals, "demo")
	c.Assert(app.TaskCount, check.Equals, 3)
	c.Assert(app.VersionCount, check.Equals, 1)
	c.Assert(len(app.Version), check.Equals, 1)

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

	// Update App
	//
	newVer := demoVersion().setName("demo").setCount(3).setCPU(0.02).setMem(10).Get()
	s.updateApp(id, newVer, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*120, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestUpdateApp() updated")

	// verify app
	app = s.inspectApp(id, c)
	c.Assert(app.Name, check.Equals, "demo")
	c.Assert(app.TaskCount, check.Equals, 3)
	c.Assert(app.VersionCount, check.Equals, 2)
	c.Assert(len(app.Version), check.Equals, 1)

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

	// Roll Back App
	//
	s.rollbackApp(id, "", c) // rollback to previous version
	err = s.waitApp(id, types.OpStatusNoop, time.Second*120, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestUpdateApp() rolled back")

	// verify app
	app = s.inspectApp(id, c)
	c.Assert(app.Name, check.Equals, "demo")
	c.Assert(app.TaskCount, check.Equals, 3)
	c.Assert(app.VersionCount, check.Equals, 2)
	c.Assert(len(app.Version), check.Equals, 1)

	// verify app versions
	vers = s.listAppVersions(id, c)
	c.Assert(len(vers), check.Equals, 2)
	c.Assert(vers[1].CPUs, check.Equals, 0.01)
	c.Assert(vers[1].Mem, check.Equals, float64(5))
	c.Assert(vers[1].Instances, check.Equals, int32(3))
	c.Assert(vers[1].RunAs, check.Equals, app.RunAs)

	// verify app tasks
	tasks = s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 3)
	for _, task := range tasks {
		c.Assert(task.Version, check.Equals, vers[1].ID)
	}

	// Remove
	//
	err = s.removeApp(id, time.Second*10, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestUpdateApp() removed")
}
