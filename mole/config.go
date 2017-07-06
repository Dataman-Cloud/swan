package mole

import (
	"errors"
	"net/url"
)

var (
	RoleMaster Role = "master"
	RoleAgent  Role = "agent"
)

type Role string

type Config struct {
	Role    Role     // both
	Listen  string   // master only
	Master  *url.URL // agent only
	Backend *url.URL // agent only
}

func (c *Config) valid() error {
	switch c.Role {
	case RoleAgent, RoleMaster:
		return nil
	default:
		return errors.New("role must be agent or master")
	}
}
