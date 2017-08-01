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

	var gitVersion = new(version.Version)
	err = s.bind(body, gitVersion)
	c.Assert(err, check.IsNil)
	c.Assert(gitVersion.GitCommit, check.Not(check.Equals), "")
	c.Assert(gitVersion.BuildTime, check.Not(check.Equals), "")
	c.Assert(gitVersion.GoVersion, check.Not(check.Equals), "")
}
