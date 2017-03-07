package janitor

import (
	"time"
)

func DefaultConfig() Config {
	config := Config{
		ListenAddr: "0.0.0.0:80",
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
	ListenAddr      string
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

type HttpHandlerCfg struct {
	FlushInterval time.Duration
	Domain        string
}

type HttpProxyServerCfg struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}
