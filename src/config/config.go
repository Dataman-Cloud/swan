package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"
)

type ManagerConfig struct {
	LogLevel           string `json:"logLevel"`
	ListenAddr         string `json:"listenAddr"`
	MesosFrameworkUser string `json:"mesosFrameworkUser"`
	Hostname           string `json:"hostname"`

	MesosURL *url.URL `json:"mesosURL"`
	ZKURL    *url.URL `json:"zkURL"`
}

type AgentConfig struct {
	DataDir       string   `json:"dataDir"`
	LogLevel      string   `json:"logLevel"`
	ListenAddr    string   `json:"listenAddr"`
	AdvertiseAddr string   `json:"advertiseAddr"`
	JoinAddrs     []string `json:"joinAddrs"`
	DNS           DNS      `json:"dns"`
	Janitor       Janitor  `json:"janitor"`
}

type DNS struct {
	Domain          string        `json:"domain"`
	RecurseOn       bool          `json:"recurseOn"`
	ListenAddr      string        `json:"listenAddr"`
	TTL             int           `json:"ttl"`
	Resolvers       []string      `json:"resolvers"`
	ExchangeTimeout time.Duration `json:"exchangeTimeout"`

	SOARname   string `json:"soarname"`
	SOAMname   string `json:"soamname"`
	SOASerial  uint32 `json:"soaserial"`
	SOARefresh uint32 `json:"soarefresh"`
	SOARetry   uint32 `json:"soaretry"`
	SOAExpire  uint32 `json:"soaexpire"`
}

type Janitor struct {
	ListenAddr    string `json:"listenAddr"`
	TLSListenAddr string `json:"tlsListenAddr"`
	TLSCertFile   string `json:"tlsCertFile"`
	TLSKeyFile    string `json:"tlsKeyFile"`
	Domain        string `json:"domain"`
	AdvertiseIP   string `json:"advertiseIP"`
}

func NewAgentConfig(c *cli.Context) (AgentConfig, error) {
	cfg := AgentConfig{
		LogLevel:   "info",
		ListenAddr: "0.0.0.0:9999",
		JoinAddrs:  []string{"0.0.0.0:9999"},

		DNS: DNS{
			Domain:     "swan.com",
			ListenAddr: "0.0.0.0:53",

			RecurseOn:       true,
			TTL:             0,
			Resolvers:       []string{"114.114.114.114"},
			ExchangeTimeout: time.Second * 3,
		},

		Janitor: Janitor{
			ListenAddr: "0.0.0.0:80",
			Domain:     "swan.com",
		},
	}

	if c.String("log-level") != "" {
		cfg.LogLevel = c.String("log-level")
	}

	if c.String("domain") != "" {
		cfg.DNS.Domain = c.String("domain")
		cfg.Janitor.Domain = c.String("domain")
	}

	if c.String("listen-addr") != "" {
		cfg.ListenAddr = c.String("listen-addr")
	}

	cfg.AdvertiseAddr = c.String("advertise-addr")
	if cfg.AdvertiseAddr == "" {
		cfg.AdvertiseAddr = cfg.ListenAddr
	}

	if c.String("gateway-advertise-ip") != "" {
		cfg.Janitor.AdvertiseIP = c.String("gateway-advertise-ip")
	}

	if c.String("gateway-listen-addr") != "" {
		cfg.Janitor.ListenAddr = c.String("gateway-listen-addr")
		if cfg.Janitor.AdvertiseIP == "" {
			cfg.Janitor.AdvertiseIP, _, _ = net.SplitHostPort(cfg.Janitor.ListenAddr)
		}
	}

	if c.String("gateway-tls-listen-addr") != "" {
		cfg.Janitor.TLSListenAddr = c.String("gateway-tls-listen-addr")
	}

	if c.String("gateway-tls-cert-file") != "" {
		cfg.Janitor.TLSCertFile = c.String("gateway-tls-cert-file")
	}

	if c.String("gateway-tls-key-file") != "" {
		cfg.Janitor.TLSKeyFile = c.String("gateway-tls-key-file")
	}

	if c.String("dns-listen-addr") != "" {
		cfg.DNS.ListenAddr = c.String("dns-listen-addr")
	}

	if c.String("dns-resolvers") != "" {
		cfg.DNS.Resolvers = strings.Split(c.String("dns-resolvers"), ",")
	}

	if ttl := c.Int("dns-ttl"); ttl > 0 {
		cfg.DNS.TTL = ttl
	}

	if c.String("join-addrs") != "" {
		cfg.JoinAddrs = strings.Split(c.String("join-addrs"), ",")
	}

	// verify tls cert/key files if gateway tls enabled
	if cfg.Janitor.TLSListenAddr != "" {
		if _, err := os.Stat(cfg.Janitor.TLSCertFile); err != nil {
			return cfg, fmt.Errorf("tsl cert file: %v", err)
		}
		if _, err := os.Stat(cfg.Janitor.TLSKeyFile); err != nil {
			return cfg, fmt.Errorf("tsl key file: %v", err)
		}
	}

	return cfg, nil
}

func NewManagerConfig(c *cli.Context) (ManagerConfig, error) {
	var err error
	cfg := ManagerConfig{
		LogLevel:           "info",
		ListenAddr:         "0.0.0.0:9999",
		MesosFrameworkUser: "root",
		Hostname:           Hostname(),
	}

	cfg.MesosURL, err = url.Parse(c.String("mesos"))
	if err != nil {
		return cfg, err
	}
	if err := validZKURL("--mesos", cfg.MesosURL); err != nil {
		return cfg, err
	}

	cfg.ZKURL, err = url.Parse(c.String("zk"))
	if err != nil {
		return cfg, err
	}
	if err := validZKURL("--zk", cfg.ZKURL); err != nil {
		return cfg, err
	}

	if c.String("listen-addr") != "" {
		cfg.ListenAddr = c.String("listen-addr")
	}

	if c.String("log-level") != "" {
		cfg.LogLevel = c.String("log-level")
	}

	return cfg, nil
}

func Hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	return hostname
}

func validZKURL(p string, zkURL *url.URL) error {
	if zkURL.Host == "" {
		return fmt.Errorf("%s not present", p)
	}

	if zkURL.Scheme != "zk" {
		return fmt.Errorf("%s scheme invalid. should be zk://", p)
	}

	if len(zkURL.Path) == 0 {
		return fmt.Errorf("no path found %s", p)
	}

	return nil
}
