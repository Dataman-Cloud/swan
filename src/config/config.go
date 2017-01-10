package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/src/utils"
	"github.com/urfave/cli"
)

type SwanConfig struct {
	LogLevel         string   `json:"log-level"`
	Mode             SwanMode `json:"mode"` // manager, agent, mixed
	DataDir          string   `json:"data-dir"`
	NoRecover        bool     `json:"no-recover"`
	Domain           string   `json:"domain"`
	SwanClusterAddrs []string `json:swan-cluster-addrs`

	Scheduler Scheduler `json:"scheduler"`
	Raft      Raft      `json:"raft"`

	DNS        DNS     `json:"dns"`
	Janitor    Janitor `json:"janitor"`
	ListenAddr string  `json:"listen-addr"`
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

type Raft struct {
	Cluster   string `json:"cluster"`
	RaftId    int    `json:"raftid"`
	StorePath string `json:"store_path"`
}

type Janitor struct {
	ListenerMode string `json:"listenerMode"`
	IP           string `json:"ip"`
	Port         int    `json:"port"`
	Domain       string `json:"domain"`
}

func NewConfig(c *cli.Context) (SwanConfig, error) {
	swanConfig := SwanConfig{
		LogLevel:         "info",
		Mode:             Mixed,
		DataDir:          "./data/",
		NoRecover:        false,
		Domain:           "swan.com",
		SwanClusterAddrs: []string{"0.0.0.0:9999"},
		ListenAddr:       "0.0.0.0:9999",

		Scheduler: Scheduler{
			ZkPath:             "0.0.0.0:2181",
			MesosFrameworkUser: "root",
			Hostname:           hostname(),
		},

		DNS: DNS{
			Domain: "swan.com",
			IP:     "0.0.0.0",
			Port:   53,

			RecurseOn:       true,
			TTL:             3,
			Resolvers:       []string{"114.114.114.114"},
			ExchangeTimeout: time.Second * 3,
		},

		Raft: Raft{
			Cluster:   "0.0.0.0:1121",
			RaftId:    1,
			StorePath: "./data/",
		},

		Janitor: Janitor{
			ListenerMode: "single_port",
			IP:           "0.0.0.0",
			Port:         80,
			Domain:       "swan.com",
		},
	}

	if c.String("log-level") != "" {
		swanConfig.LogLevel = c.String("log-level")
	}

	if c.String("mode") != "" {
		if utils.SliceContains([]string{"mixed", "manager", "agent"}, c.String("mode")) {
			swanConfig.Mode = SwanMode(c.String("mode"))
		} else {
			return swanConfig, errors.New("mode should be one of mixed, manager or agent")
		}
	}

	if c.String("data-dir") != "" {
		swanConfig.DataDir = c.String("data-dir")
		if !strings.HasSuffix(swanConfig.DataDir, "/") {
			swanConfig.DataDir = swanConfig.DataDir + "/"
		}
	}

	if c.String("zk-path") != "" {
		swanConfig.Scheduler.ZkPath = c.String("zk-path")
	}

	if c.String("raft-cluster") != "" {
		swanConfig.Raft.Cluster = c.String("raft-cluster")
	}

	if c.Int("raftid") != 0 {
		swanConfig.Raft.RaftId = c.Int("raftid")
		swanConfig.DataDir = fmt.Sprintf(swanConfig.DataDir+"%d", swanConfig.Raft.RaftId)
		swanConfig.Raft.StorePath = swanConfig.DataDir
	}

	if c.String("cluster-addrs") != "" {
		swanConfig.SwanClusterAddrs = strings.Split(c.String("cluster-addrs"), ",")
		swanConfig.ListenAddr = swanConfig.SwanClusterAddrs[swanConfig.Raft.RaftId-1]
	}

	if c.String("domain") != "" {
		swanConfig.Domain = c.String("domain")
		swanConfig.DNS.Domain = c.String("domain")
		swanConfig.Janitor.Domain = c.String("domain")
	}

	// TODO(upccup): this is not the optimal solution. Maybe we can use listen-addr replace --swan-cluster
	// if swan mode is mixed agent listen addr use the same with manager. But if the mode is agent we need
	// a listen-addr just for agent
	if swanConfig.Mode == Agent {
		swanConfig.ListenAddr = c.String("listen-addr")
	}

	return swanConfig, nil
}

func hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	return hostname
}
