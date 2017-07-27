package main

import (
	"net/http"

	check "gopkg.in/check.v1"

	"github.com/Dataman-Cloud/swan/version"
)

func (s *ApiSuite) TestGetVersion(c *check.C) {
	code, body, err := s.sendRequest("GET", "/version", nil)
	c.Assert(err, check.IsNil)
	c.Assert(code, check.Equals, http.StatusOK)

	err = s.bind(body, new(version.Version))
	c.Assert(err, check.IsNil)
}
