package util

import (
	"errors"
	"os"
	"time"

	"github.com/urfave/cli"
)

type SwanConfig struct {
	LogLevel   string `json:"log-level"`
	Mode       string `json:"manager"` // manager, agent, mixed
	Standalone bool   `json:"standalone"`

	Scheduler    Scheduler    `json:"scheduler"`
	DNS          DNS          `json:"dns"`
	HttpListener HttpListener `json:"httpListener"`
	IPAM         IPAM         `json:"ipam"`
	Raft         Raft         `json:"raft"`
}

type Scheduler struct {
	MesosMasters           []string `json:"mesos-masters"`
	MesosFrameworkUser     string   `json:"mesos-framwork-user"`
	Hostname               string   `json:"hostname"`
	EnableLocalHealthcheck bool     `json:"local-healthcheck"`
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
	TCPAddr  string `json:"tcp-addr"`
	UnixAddr string `json:"unix-addr"`
}

type IPAM struct {
	StorePath string `json:"store_path"`
}

type Raft struct {
	Cluster   string `json:"cluster"`
	RaftId    int    `json:"raftid"`
	StorePath string `json:"store_path"`
}

func NewConfig(c *cli.Context) (SwanConfig, error) {
	config := SwanConfig{
		LogLevel:   c.String("log-level"),
		Mode:       c.String("mode"),
		Standalone: c.Bool("standablone"),
		HttpListener: HttpListener{
			TCPAddr:  c.String("addr"),
			UnixAddr: c.String("unix_addr"),
		},

		Scheduler: Scheduler{
			MesosMasters:           []string{c.String("masters")},
			MesosFrameworkUser:     c.String("user"),
			Hostname:               hostname(),
			EnableLocalHealthcheck: c.Bool("local-healthcheck"),
		},

		DNS: DNS{
			EnableDnsProxy: c.Bool("enable-dns-proxy"),
			Domain:         "swan",
			RecurseOn:      true,
			Listener:       "0.0.0.0",
			Port:           53,
			Resolvers:      []string{"114.114.114.114"},
		},

		IPAM: IPAM{
			StorePath: c.String("work-dir"),
		},

		Raft: Raft{
			Cluster:   c.String("cluster"),
			RaftId:    c.Int("raftid"),
			StorePath: c.String("work-dir"),
		},
	}

	return validateAndFormatConfig(config)
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
		if os.Getuid() == 0 || (len(os.Getenv("SUDO_UID")) > 0) {
			return config, errors.New("no permission to run DNS server, run as root or sudoer")
		}
	}
	return config, nil
}
