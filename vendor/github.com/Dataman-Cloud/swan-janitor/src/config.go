package janitor

import (
	"net"
	"time"
)

func DefaultConfig() Config {
	ip := net.ParseIP("0.0.0.0").String()

	config := Config{
		Listener: ListenerCfg{
			IP:          ip,
			DefaultPort: "80",
		},
		HttpHandler: HttpHandlerCfg{
			FlushInterval: time.Second * 1,
			Domain:        "lvh.me",
		},
		HttpProxyServer: HttpProxyServerCfg{
			ReadTimeout:  time.Second * 1,
			WriteTimeout: time.Second * 1,
		},
	}

	return config
}

type Config struct {
	Proxy           ProxyCfg
	Listener        ListenerCfg
	HttpHandler     HttpHandlerCfg
	HttpProxyServer HttpProxyServerCfg
}

type ProxyCfg struct {
	Strategy              string
	Matcher               string
	NoRouteStatus         int
	MaxConn               int
	ShutdownWait          time.Duration
	DialTimeout           time.Duration
	ResponseHeaderTimeout time.Duration
	KeepAliveTimeout      time.Duration
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	FlushInterval         time.Duration
	LocalIP               string
	ClientIPHeader        string
	TLSHeader             string
	TLSHeaderValue        string
}

type ListenerCfg struct {
	IP          string
	DefaultPort string
}

type HttpHandlerCfg struct {
	FlushInterval time.Duration
	Domain        string
}

type HttpProxyServerCfg struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}
