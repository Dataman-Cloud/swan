package config

import (
	"fmt"
	"net/url"

	"github.com/urfave/cli"
)

type ManagerConfig struct {
	LogLevel   string `json:"logLevel"`
	Listen     string `json:"listenAddr"`
	EnableCORS bool

	MesosURL *url.URL `json:"mesosURL"`
	ZKURL    *url.URL `json:"zkURL"`
	Strategy string   `json:"strategy"`

	ReconciliationInterval  float64 `json:"reconciliationInterval"`
	ReconciliationStep      int64   `json:"reconciliationStep"`
	ReconciliationStepDelay float64 `json:"reconciliationStepDelay"`
}

func NewManagerConfig(c *cli.Context) (*ManagerConfig, error) {
	cfg := &ManagerConfig{
		LogLevel: "info",
		Listen:   "0.0.0.0:9999",
	}

	var err error

	cfg.MesosURL, err = url.Parse(c.String("mesos"))
	if err != nil {
		return cfg, err
	}

	cfg.ZKURL, err = url.Parse(c.String("zk"))
	if err != nil {
		return cfg, err
	}

	if c.String("listen") != "" {
		cfg.Listen = c.String("listen")
	}

	if c.String("log-level") != "" {
		cfg.LogLevel = c.String("log-level")
	}

	if c.String("strategy") != "" {
		cfg.Strategy = c.String("strategy")
	}

	if c.Float64("reconciliation-interval") != 0 {
		cfg.ReconciliationInterval = c.Float64("reconciliation-interval")
	}

	if c.Int64("reconciliation-step") != 0 {
		cfg.ReconciliationStep = c.Int64("reconciliation-step")
	}

	if c.Float64("reconciliation-step-delay") != 0 {
		cfg.ReconciliationStepDelay = c.Float64("reconciliation-step-delay")
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *ManagerConfig) validate() error {
	if c.ZKURL.Host == "" {
		return fmt.Errorf("zk host can not be empty")
	}

	if c.ZKURL.Scheme != "zk" {
		return fmt.Errorf("malformed scheme for zk url.")
	}

	if c.ZKURL.Path == "" {
		return fmt.Errorf("zk url not corrected. path must be provied")
	}

	if c.Strategy != "random" && c.Strategy != "spread" && c.Strategy != "binpack" {
		return fmt.Errorf("strategy not supported. must be one of the 'random, spread, binpack'")
	}

	if c.ReconciliationInterval <= 0 {
		return fmt.Errorf("reconciliation interval must be positive")
	}

	if c.ReconciliationStep <= 0 {
		return fmt.Errorf("reconciliation step must be positive")
	}

	if c.ReconciliationStepDelay <= 0 {
		return fmt.Errorf("reconciliation step delay must be positive")
	}

	return nil
}
