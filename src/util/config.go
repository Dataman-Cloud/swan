package util

import (
	"encoding/json"
	//"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

type SwanConfig struct {
	LogLevel   string `json:"log-level"`
	Mode       string `json:"manager"` // manager, agent, mixed
	Standalone bool   `json:"standalone"`
	DataDir    string `json:"data-dir"`

	Scheduler    Scheduler    `json:"scheduler"`
	DNS          DNS          `json:"dns"`
	HttpListener HttpListener `json:"httpListener"`
	IPAM         IPAM         `json:"ipam"`
	Raft         Raft         `json:"raft"`
	SwanCluster  []string     `json:swanCluster`
}

type Scheduler struct {
	MesosMasters           string `json:"mesos-masters"`
	MesosFrameworkUser     string `json:"mesos-framwork-user"`
	Hostname               string `json:"hostname"`
	EnableLocalHealthcheck bool   `json:"local-healthcheck"`
	HttpAddr               string
}

type DNS struct {
	EnableDnsProxy bool `json:"enable-dns-proxy"`

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
	TCPAddr  string `json:"addr"`
	UnixAddr string `json:"sock"`
}

type IPAM struct {
	StorePath string `json:"store_path"`
}

type Raft struct {
	Cluster   string `json:"cluster"`
	RaftId    int    `json:"raftid"`
	StorePath string `json:"store_path"`
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
	if c.Bool("standalone") != false {
		swanConfig.Standalone = c.Bool("standalone")
	}
	if c.String("cluster") != "" {
		swanConfig.SwanCluster = strings.Split(c.String("cluster"), ",")
	}
	if c.String("sock") != "" {
		swanConfig.HttpListener.UnixAddr = c.String("sock")
	}
	if c.String("master") != "" {
		swanConfig.Scheduler.MesosMasters = c.String("master")
	}
	if c.String("user") != "" {
		swanConfig.Scheduler.MesosFrameworkUser = c.String("user")
	}
	if c.Bool("local-healthcheck") != false {
		swanConfig.Scheduler.EnableLocalHealthcheck = c.Bool("local-healthcheck")
	}

	if c.Bool("enable-dns-proxy") != false {
		swanConfig.DNS.EnableDnsProxy = c.Bool("enable-dns-proxy")
		swanConfig.DNS.ExchangeTimeout = time.Second * 3
	}

	if c.String("raft-cluster") != "" {
		swanConfig.Raft.Cluster = c.String("raft-cluster")
	}
	if c.Int("raftid") != 0 {
		swanConfig.Raft.RaftId = c.Int("raftid")
	}

	swanConfig.IPAM.StorePath = swanConfig.DataDir
	swanConfig.Raft.StorePath = swanConfig.DataDir

	swanConfig.HttpListener.TCPAddr = swanConfig.SwanCluster[swanConfig.Raft.RaftId-1]
	swanConfig.Scheduler.HttpAddr = swanConfig.HttpListener.TCPAddr

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
	if config.DNS.EnableDnsProxy {
		//if os.Getuid() == 0 || (len(os.Getenv("SUDO_UID")) > 0) {
		//return config, errors.New("no permission to run DNS server, run as root or sudoer")
		//}
	}
	return config, nil
}
