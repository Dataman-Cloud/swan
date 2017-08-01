package main

import (
	"net/http"

	check "gopkg.in/check.v1"
)

func (s *ApiSuite) TestPing(c *check.C) {
	code, body, err := s.sendRequest("GET", "/ping", nil)
	c.Assert(err, check.IsNil)
	c.Assert(code, check.Equals, http.StatusOK)
	c.Assert(string(body), check.Equals, `"pong"`)
}
