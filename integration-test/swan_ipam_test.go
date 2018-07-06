package main

import (
	"fmt"
	"time"

	"github.com/Dataman-Cloud/swan/agent/ipam"
	"github.com/Dataman-Cloud/swan/types"

	check "gopkg.in/check.v1"
)

func (s *ApiSuite) TestIPAM(c *check.C) {
	// Purge
	//
	err := s.purge(time.Second*30, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestCreateApp() purged")

	// get all agents
	agents := s.listAgents(c)

	// create network on one of the agents
	//

	type Config struct {
		Subnet  string
		Gateway string
	}

	type IPAM struct {
		Driver  string
		Config  []*Config
		Options map[string]string
	}

	var body = struct {
		Name           string
		CheckDuplicate bool
		EnableIPv6     bool
		IPAM           *IPAM
		Internal       bool
		Options        map[string]string
		Labels         map[string]string
	}{
		Name:           "swan",
		CheckDuplicate: true,
		IPAM: &IPAM{
			Driver: "swan",
			Config: []*Config{
				{
					Subnet:  "10.0.0.0/24",
					Gateway: "10.0.0.1",
				},
			},
			Options: nil,
		},
		Internal: false,
		Options: map[string]string{
			"parent": "ens224",
		},
		Labels: nil,
	}

	// delete network swan first
	err = s.deleteNetwork(agents[0], "swan", c)
	c.Assert(err, check.IsNil)

	// create network swan
	id := s.createNetwork(agents[0], body, c)

	// verify
	network := s.inspectNetwork(agents[0], id, c)
	c.Assert(network, check.Equals, "swan")

	// create subnet
	pool := &ipam.IPPoolRange{
		IPStart: "10.0.0.2/24",
		IPEnd:   "10.0.0.20/24",
	}
	err = s.createSubnet(agents[0], pool, c)
	c.Assert(err, check.IsNil)

	// Purge
	//
	err = s.purge(time.Second*30, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestCreateApp() purged")

	// New Create App
	//
	ver := demoVersion().setName("demo").setCount(5).setCPU(0.01).setMem(5).Get()
	id = s.createApp(ver, c)
	err = s.waitApp(id, types.OpStatusNoop, time.Second*30, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestCreateApp() created")

	// verify tasks ip
	tasks := s.listAppTasks(id, c)
	c.Assert(len(tasks), check.Equals, 5)
	var i int
	for _, task := range tasks {
		if task.IP != "" {
			i++
		}
	}

	c.Assert(i, check.Equals, 5)

	// Remove
	//
	err = s.removeApp(id, time.Second*10, c)
	c.Assert(err, check.IsNil)
	fmt.Println("TestCreateApp() removed")
}
