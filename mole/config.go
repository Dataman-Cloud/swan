package mole

import (
	"errors"
	"net/url"
	"os"
)

var (
	roleMaster role = "master"
	roleAgent  role = "agent"
)

type role string

type Config struct {
	role    role     // both
	listen  string   // master only
	master  *url.URL // agent only
	backend *url.URL // agent only
}

func (c *Config) valid() error {
	switch c.role {

	case roleAgent:
		if c.master == nil {
			return errors.New("malform master endpoint")
		}
		if c.backend == nil {
			return errors.New("malform backend endpoint")
		}

	case roleMaster:
		if c.listen == "" {
			return errors.New("malform listen address")
		}

	default:
		return errors.New("role must be agent or master")
	}

	return nil
}

func ConfigFromEnv() (*Config, error) {
	cfg := &Config{
		role:   role(os.Getenv("MOLE_ROLE")),
		listen: os.Getenv("MOLE_LISTEN"),
	}
	if burl, err := url.Parse(os.Getenv("MOLE_BACKEND_ENDPOINT")); err == nil {
		cfg.backend = burl
	}
	if murl, err := url.Parse(os.Getenv("MOLE_MASTER_ENDPOINT")); err == nil {
		cfg.master = murl
	}
	if err := cfg.valid(); err != nil {
		return nil, err
	}
	return cfg, nil
}
