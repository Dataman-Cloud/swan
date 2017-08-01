package main

import (
	"net/http"

	check "gopkg.in/check.v1"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *ApiSuite) TestGetLeader(c *check.C) {
	code, body, err := s.sendRequest("GET", "/v1/leader", nil)
	c.Assert(err, check.IsNil)
	c.Assert(code, check.Equals, http.StatusOK)

	var leader = new(types.Leader)
	err = s.bind(body, leader)
	c.Assert(err, check.IsNil)
	c.Assert(leader.Leader, check.Not(check.Equals), "")
}
