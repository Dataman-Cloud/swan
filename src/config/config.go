package config

import (
	"encoding/json"
	"fmt"

	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

type SwanConfig struct {
	LogLevel  string `json:"log-level"`
	Mode      string `json:"manager"` // manager, agent, mixed
	DataDir   string `json:"data-dir"`
	NoRecover bool   `json:"no-recover"`

	Scheduler    Scheduler    `json:"scheduler"`
	DNS          DNS          `json:"dns"`
	HttpListener HttpListener `json:"httpListener"`
	Raft         Raft         `json:"raft"`
	SwanCluster  []string     `json:swanCluster`

	Janitor Janitor `json:"janitor"`
}

type Scheduler struct {
	ZkUrl              string `json:"zkurl"`
	MesosFrameworkUser string `json:"mesos-framwork-user"`
	Hostname           string `json:"hostname"`
	UnixAddr           string
}

type DNS struct {
	EnableDns bool `json:"enable-dns"`

	Domain    string `json:"domain"`
	RecurseOn bool   `json:"recurse_on"`
	Listener  string `json:"ip"`
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

type HttpListener struct {
	TCPAddr string `json:"addr"`
}

type Raft struct {
	Cluster   string `json:"cluster"`
	RaftId    int    `json:"raftid"`
	StorePath string `json:"store_path"`
}

type Janitor struct {
	EnableProxy  bool   `json:"enableProxy"`
	ListenerMode string `json:"listenerMode"`
	IP           string `json:"ip"`
	Port         string `json:"port"`
	Domain       string `json:"domain"`
}

func LoadConfig(configFile string) SwanConfig {
	var swanConfig SwanConfig
	logrus.Debug("configfile: ", configFile)
	config, err := ioutil.ReadFile(configFile)
	if err != nil {
		logrus.Errorf("Failed to read config file %s: %s", configFile, err.Error())
		return swanConfig
	}
	err = json.Unmarshal(config, &swanConfig)
	if err != nil {
		logrus.Errorf("Failed to unmarshal configs from configFile %s:%s", configFile, err.Error())
	}
	return swanConfig
}

func NewConfig(c *cli.Context) (SwanConfig, error) {
	configFile := c.String("config-file")
	swanConfig := LoadConfig(configFile)
	swanConfig.Scheduler.Hostname = hostname()
	if c.String("log-level") != "" {
		swanConfig.LogLevel = c.String("log-level")
	}
	if c.String("mode") != "" {
		swanConfig.Mode = c.String("mode")
	}

	if c.String("data-dir") != "" {
		swanConfig.DataDir = c.String("data-dir")
		if !strings.HasSuffix(swanConfig.DataDir, "/") {
			swanConfig.DataDir = swanConfig.DataDir + "/"
		}
	}

	if c.String("no-recover") != "" {
		swanConfig.NoRecover = c.Bool("no-recover")
	}

	if c.String("cluster") != "" {
		swanConfig.SwanCluster = strings.Split(c.String("cluster"), ",")
	}

	if c.String("zk") != "" {
		swanConfig.Scheduler.ZkUrl = c.String("zk")
	}

	if c.String("raft-cluster") != "" {
		swanConfig.Raft.Cluster = c.String("raft-cluster")
	}
	if c.Int("raftid") != 0 {
		swanConfig.Raft.RaftId = c.Int("raftid")
	}

	swanConfig.Janitor.EnableProxy = c.Bool("enable-proxy")
	swanConfig.DNS.EnableDns = c.Bool("enable-dns")

	swanConfig.HttpListener.TCPAddr = swanConfig.SwanCluster[swanConfig.Raft.RaftId-1]
	swanConfig.Scheduler.MesosFrameworkUser = "root"
	swanConfig.DNS.ExchangeTimeout = time.Second * 3

	// ugly code maybe we should use raft id as node id and merge raft.StorePath to DataDir
	swanConfig.DataDir = fmt.Sprintf(swanConfig.DataDir+"%d/", swanConfig.Raft.RaftId)
	swanConfig.Raft.StorePath = swanConfig.DataDir

	return validateAndFormatConfig(swanConfig)
}

func hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	return hostname
}

func validateAndFormatConfig(config SwanConfig) (c SwanConfig, e error) {
	return config, nil
}
