package config

import (
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"
)

type ManagerConfig struct {
	LogLevel          string   `json:"logLevel"`
	DataDir           string   `json:"dataDir"`
	RaftAdvertiseAddr string   `json:"raftAdvertiseAddr"`
	RaftListenAddr    string   `json:"raftListenAddr"`
	ListenAddr        string   `json:"listenAddr"`
	AdvertiseAddr     string   `json:"advertiseAddr"`
	JoinAddrs         []string `json:"joinAddrs"`

	Scheduler Scheduler `json:"scheduler"`
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

type Scheduler struct {
	ZkPath             string `json:"zkpath"`
	MesosFrameworkUser string `json:"mesos-framwork-user"`
	Hostname           string `json:"hostname"`
}

type DNS struct {
	Domain    string `json:"domain"`
	RecurseOn bool   `json:"recurse_on"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`

	SOARname   string `json:"soarname"`
	SOAMname   string `json:"soamname"`
	SOASerial  uint32 `json:"soaserial"`
	SOARefresh uint32 `json:"soarefresh"`
	SOARetry   uint32 `json:"soaretry"`
	SOAExpire  uint32 `json:"soaexpire"`

	TTL int `json:"ttl"`

	Resolvers       []string      `json:"resolvers"`
	ExchangeTimeout time.Duration `json:"exchange_timeout"`
}

type Janitor struct {
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Domain      string `json:"domain"`
	AdvertiseIP string `json:"advertiseIp"`
}

func NewAgentConfig(c *cli.Context) AgentConfig {
	agentConfig := AgentConfig{
		LogLevel:   "info",
		DataDir:    "./data/",
		ListenAddr: "0.0.0.0:9999",
		JoinAddrs:  []string{"0.0.0.0:9999"},

		DNS: DNS{
			Domain: "swan.com",
			IP:     "0.0.0.0",
			Port:   53,

			RecurseOn:       true,
			TTL:             3,
			Resolvers:       []string{"114.114.114.114"},
			ExchangeTimeout: time.Second * 3,
		},

		Janitor: Janitor{
			IP:          "0.0.0.0",
			Port:        80,
			AdvertiseIP: "0.0.0.0",
			Domain:      "swan.com",
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

	if c.String("janitor-advertise-ip") != "" {
		agentConfig.Janitor.AdvertiseIP = c.String("janitor-advertise-ip")
	}

	if c.String("join-addrs") != "" {
		agentConfig.JoinAddrs = strings.Split(c.String("join-addrs"), ",")
	}

	return agentConfig
}

func NewManagerConfig(c *cli.Context) ManagerConfig {
	managerConfig := ManagerConfig{
		LogLevel:       "info",
		DataDir:        "./data/",
		ListenAddr:     "0.0.0.0:9999",
		RaftListenAddr: "0.0.0.0:2111",
		JoinAddrs:      []string{"0.0.0.0:9999"},

		Scheduler: Scheduler{
			ZkPath:             "0.0.0.0:2181",
			MesosFrameworkUser: "root",
			Hostname:           Hostname(),
		},
	}

	if c.String("log-level") != "" {
		managerConfig.LogLevel = c.String("log-level")
	}

	if c.String("data-dir") != "" {
		managerConfig.DataDir = c.String("data-dir")
		if !strings.HasSuffix(managerConfig.DataDir, "/") {
			managerConfig.DataDir = managerConfig.DataDir + "/"
		}
	}

	if c.String("zk-path") != "" {
		managerConfig.Scheduler.ZkPath = c.String("zk-path")
	}

	if c.String("listen-addr") != "" {
		managerConfig.ListenAddr = c.String("listen-addr")
	}

	managerConfig.AdvertiseAddr = c.String("advertise-addr")
	if managerConfig.AdvertiseAddr == "" {
		managerConfig.AdvertiseAddr = managerConfig.ListenAddr
	}

	if c.String("raft-advertise-addr") != "" {
		managerConfig.RaftAdvertiseAddr = c.String("raft-advertise-addr")
	}

	if c.String("raft-listen-addr") != "" {
		managerConfig.RaftListenAddr = c.String("raft-listen-addr")
	}

	if managerConfig.RaftAdvertiseAddr == "" {
		managerConfig.RaftAdvertiseAddr = managerConfig.RaftListenAddr
	}

	if c.String("join-addrs") != "" {
		managerConfig.JoinAddrs = strings.Split(c.String("join-addrs"), ",")
	}

	SchedulerConfig = managerConfig.Scheduler

	return managerConfig
}

func Hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	return hostname
}

var SchedulerConfig Scheduler
