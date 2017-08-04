package main

import (
	"fmt"
	"strings"
	"time"

	check "gopkg.in/check.v1"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *ApiSuite) TestRunCompose(c *check.C) {
	// Purge
	//
	err := s.purge(time.Second*30, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestRunCompose() purged")

	// New Create Compose
	//
	yaml, exts := demoYAML()
	fmt.Printf("launch compose with yaml text:\n%s\n", yaml)
	cmp := demoCompose().setName("demo").setDesc("desc").setYAML(yaml, exts).Get()
	id := s.runCompose(cmp, c)
	err = s.waitCompose(id, types.OpStatusNoop, time.Second*60, c)
	c.Assert(err, check.IsNil)
	c.Assert(cmp.ErrMsg, check.Equals, "")
	fmt.Println("TestRunCompose() created")

	// verify app
	wrapper := s.inspectCompose(id, c)
	c.Assert(wrapper.Name, check.Equals, "demo")
	c.Assert(wrapper.Desc, check.Equals, "desc")
	for _, app := range wrapper.Apps {
		c.Assert(strings.HasSuffix(app.ID, wrapper.DisplayName), check.Equals, true)
		c.Assert(app.OpStatus, check.Equals, types.OpStatusNoop)
		c.Assert(app.VersionCount, check.Equals, 1)
		c.Assert(app.ErrMsg, check.Equals, "")
	}

	// Remove
	//
	err = s.removeCompose(id, time.Second*60, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestRunCompose() removed")
}
