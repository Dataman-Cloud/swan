package config

import (
	"errors"
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

	MesosZkPath *url.URL `json:"mesosZkPath"`
	ZkPath      *url.URL `json:"zkPath"`
}

type AgentConfig struct {
	DataDir          string   `json:"dataDir"`
	LogLevel         string   `json:"logLevel"`
	ListenAddr       string   `json:"listenAddr"`
	AdvertiseAddr    string   `json:"advertiseAddr"`
	GossipListenAddr string   `json:"gossipListenAddr"`
	GossipJoinAddr   string   `json:"gossipJoinAddr"`
	JoinAddrs        []string `json:"joinAddrs"`
	DNS              DNS      `json:"dns"`
	Janitor          Janitor  `json:"janitor"`
}

type DNS struct {
	Domain     string `json:"domain"`
	RecurseOn  bool   `json:"recurseOn"`
	ListenAddr string `json:"listenAddr"`

	SOARname   string `json:"soarname"`
	SOAMname   string `json:"soamname"`
	SOASerial  uint32 `json:"soaserial"`
	SOARefresh uint32 `json:"soarefresh"`
	SOARetry   uint32 `json:"soaretry"`
	SOAExpire  uint32 `json:"soaexpire"`

	TTL int `json:"ttl"`

	Resolvers       []string      `json:"resolvers"`
	ExchangeTimeout time.Duration `json:"exchangeTimeout"`
}

type Janitor struct {
	ListenAddr  string `json:"listenAddr"`
	Domain      string `json:"domain"`
	AdvertiseIP string `json:"ddvertiseIP"`
}

func NewAgentConfig(c *cli.Context) AgentConfig {
	agentConfig := AgentConfig{
		LogLevel:         "info",
		DataDir:          "./data/",
		ListenAddr:       "0.0.0.0:9999",
		JoinAddrs:        []string{"0.0.0.0:9999"},
		GossipListenAddr: "0.0.0.0:5000",

		DNS: DNS{
			Domain:     "swan.com",
			ListenAddr: "0.0.0.0:53",

			RecurseOn:       true,
			TTL:             3,
			Resolvers:       []string{"114.114.114.114"},
			ExchangeTimeout: time.Second * 3,
		},

		Janitor: Janitor{
			ListenAddr: "0.0.0.0:80",
			Domain:     "swan.com",
		},
	}

	if c.String("log-level") != "" {
		agentConfig.LogLevel = c.String("log-level")
	}

	if c.String("data-dir") != "" {
		agentConfig.DataDir = c.String("data-dir")
		if !strings.HasSuffix(agentConfig.DataDir, "/") {
			agentConfig.DataDir = agentConfig.DataDir + "/"
		}
	}

	if c.String("domain") != "" {
		agentConfig.DNS.Domain = c.String("domain")
		agentConfig.Janitor.Domain = c.String("domain")
	}

	if c.String("listen-addr") != "" {
		agentConfig.ListenAddr = c.String("listen-addr")
	}

	agentConfig.AdvertiseAddr = c.String("advertise-addr")
	if agentConfig.AdvertiseAddr == "" {
		agentConfig.AdvertiseAddr = agentConfig.ListenAddr
	}

	if c.String("gateway-advertise-ip") != "" {
		agentConfig.Janitor.AdvertiseIP = c.String("gateway-advertise-ip")
	}

	if c.String("gateway-listen-addr") != "" {
		agentConfig.Janitor.ListenAddr = c.String("gateway-listen-addr")

		if agentConfig.Janitor.AdvertiseIP == "" {
			agentConfig.Janitor.AdvertiseIP, _, _ = net.SplitHostPort(agentConfig.Janitor.ListenAddr)
		}
	}

	if c.String("dns-listen-addr") != "" {
		agentConfig.DNS.ListenAddr = c.String("dns-listen-addr")
	}

	if c.String("dns-resolvers") != "" {
		agentConfig.DNS.Resolvers = strings.Split(c.String("dns-resolvers"), ",")
	}

	if c.String("join-addrs") != "" {
		agentConfig.JoinAddrs = strings.Split(c.String("join-addrs"), ",")
	}

	if c.String("gossip-listen-addr") != "" {
		agentConfig.GossipListenAddr = c.String("gossip-listen-addr")
	}

	if c.String("gossip-join-addr") != "" {
		agentConfig.GossipJoinAddr = c.String("gossip-join-addr")
	}

	return agentConfig
}

func NewManagerConfig(c *cli.Context) (ManagerConfig, error) {
	var err error
	managerConfig := ManagerConfig{
		LogLevel:           "info",
		ListenAddr:         "0.0.0.0:9999",
		MesosFrameworkUser: "root",
		Hostname:           Hostname(),
	}

	managerConfig.MesosZkPath, err = url.Parse(c.String("mesos-zk-path"))
	if err != nil {
		return managerConfig, err
	}
	if err := validZkURL("MesosZkPath", managerConfig.MesosZkPath); err != nil {
		return managerConfig, err
	}

	managerConfig.ZkPath, err = url.Parse(c.String("zk-path"))
	if err != nil {
		return managerConfig, err
	}
	if err := validZkURL("ZkPath", managerConfig.ZkPath); err != nil {
		return managerConfig, err
	}

	if (managerConfig.MesosZkPath.Path == managerConfig.ZkPath.Path) &&
		(managerConfig.MesosZkPath.Host == managerConfig.ZkPath.Host) {
		return managerConfig, errors.New("ZkPath shouldn't be same as MesosZkPath")
	}

	if c.String("listen-addr") != "" {
		managerConfig.ListenAddr = c.String("listen-addr")
	}

	if c.String("log-level") != "" {
		managerConfig.LogLevel = c.String("log-level")
	}

	return managerConfig, nil
}

func Hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	return hostname
}

func validZkURL(which string, zkUrl *url.URL) error {
	if zkUrl.Host == "" {
		return errors.New(fmt.Sprintf("%s not present", which))
	}

	if zkUrl.Scheme != "zk" {
		return errors.New(fmt.Sprintf("%s should have valid scheme, default should be zk://", which))
	}

	if len(zkUrl.Path) == 0 {
		return errors.New(fmt.Sprintf("%s should provide meaningful path.   eg. swan", which))
	}

	return nil
}
