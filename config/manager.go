package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/urfave/cli"
)

type ManagerConfig struct {
	LogLevel   string `json:"logLevel"`
	Listen     string `json:"listenAddr"`
	EnableCORS bool

	MesosURL *url.URL `json:"mesosURL"` // mesos zk url

	StoreType string   `json:"store_type"` // db store type
	ZKURL     *url.URL `json:"zkURL"`      // zk store url (currently we always require this for HA)
	EtcdAddrs []string `json:"etcd_addrs"` // etcd store addrs

	Strategy string `json:"strategy"`

	ReconciliationInterval  float64 `json:"reconciliationInterval"`
	ReconciliationStep      int64   `json:"reconciliationStep"`
	ReconciliationStepDelay float64 `json:"reconciliationStepDelay"`
	HeartbeatTimeout        float64 `json:"heartbeatTimeout"`
	MaxTasksPerOffer        int     `json:"maxTasksPerOffer"`
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

	if c.String("store-type") != "" {
		cfg.StoreType = c.String("store-type")
	}

	cfg.ZKURL, err = url.Parse(c.String("zk"))
	if err != nil {
		return cfg, err
	}

	if c.String("etcd-addrs") != "" {
		cfg.EtcdAddrs = strings.Split(c.String("etcd-addrs"), ",")
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

	if c.Float64("heartbeat-timeout") != 0 {
		cfg.HeartbeatTimeout = c.Float64("heartbeat-timeout")
	}

	if max := c.Int("max-tasks-per-offer"); max != 0 {
		cfg.MaxTasksPerOffer = max
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *ManagerConfig) validate() error {

	if typ := c.StoreType; typ == "etcd" && len(c.EtcdAddrs) == 0 {
		return fmt.Errorf("at least one of etcd cluster address required")
	}

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
