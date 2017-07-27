package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"gopkg.in/check.v1"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"
)

func (s *ApiSuite) TestCreateApp(c *check.C) {
	var (
		name  = utils.RandomString(8)
		count = 20
	)

	ver := newVersion(name, count)
	data, _ := s.encode(ver)
	fmt.Print("POST http://" + s.SwanHost + "/v1/apps")
	code, body, err := s.sendRequest("POST", "/v1/apps", bytes.NewReader(data))
	c.Assert(err, check.IsNil)
	c.Assert(code, check.Equals, http.StatusCreated)
	fmt.Println(" ", code)

	var rt struct {
		Id string
	}

	s.bind(body, &rt)

	c.Assert(rt.Id, check.Equals, name+".default.integration.unnamed")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fmt.Println("Waiting for app", rt.Id, "to be running")
	for {
		select {
		case <-time.After(5 * time.Minute):
			c.Errorf("task count not exceed %d after 5 minute", count)
			return
		case <-ticker.C:
			fmt.Print(".")
			code, body, err := s.sendRequest("GET", "/v1/apps/"+rt.Id, nil)
			c.Assert(err, check.IsNil)
			c.Assert(code, check.Equals, http.StatusOK)

			app := new(types.Application)

			s.bind(body, app)

			if app.TaskCount == count {
				fmt.Println("OK")
				goto DELETE
			}
		}
	}

DELETE:
	fmt.Print("DELETE http://" + s.SwanHost + "/v1/apps/" + rt.Id)
	code, _, err = s.sendRequest("DELETE", "/v1/apps/"+rt.Id, nil)
	c.Assert(err, check.IsNil)
	c.Assert(code, check.Equals, http.StatusNoContent)
	fmt.Println(" ", code)

	fmt.Println("Waiting for app", rt.Id, "to be killed")
	for {
		select {
		case <-time.After(5 * time.Minute):
			c.Errorf("app %s not deleted after 5 minute", rt.Id)
			return
		case <-ticker.C:
			fmt.Print(".")
			code, _, err := s.sendRequest("GET", "/v1/apps/"+rt.Id, nil)
			c.Assert(err, check.IsNil)
			if code == http.StatusNotFound {
				fmt.Println("OK")
				return
			}
		}
	}
}
