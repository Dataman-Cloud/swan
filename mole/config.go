package mole

import (
	"errors"
	"net/url"
	"os"
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

	case RoleAgent:
		if c.Master == nil {
			return errors.New("malform master endpoint")
		}
		if c.Backend == nil {
			return errors.New("malform backend endpoint")
		}

	case RoleMaster:
		if c.Listen == "" {
			return errors.New("malform listen address")
		}

	default:
		return errors.New("role must be agent or master")
	}

	return nil
}

func ConfigFromEnv() (*Config, error) {
	cfg := &Config{
		Role:   Role(os.Getenv("MOLE_ROLE")),
		Listen: os.Getenv("MOLE_LISTEN"),
	}
	if burl, err := url.Parse(os.Getenv("MOLE_BACKEND_ENDPOINT")); err == nil {
		cfg.Backend = burl
	}
	if murl, err := url.Parse(os.Getenv("MOLE_MASTER_ENDPOINT")); err == nil {
		cfg.Master = murl
	}
	if err := cfg.valid(); err != nil {
		return nil, err
	}
	return cfg, nil
}
