package config

import (
	"net"
	"net/http"
	"time"
)

const (
	SINGLE_LISTENER_MODE    = "single_port"
	MULTIPORT_LISTENER_MODE = "multi_port"
)

func DefaultConfig() Config {
	ip := net.ParseIP("0.0.0.0").String()

	config := Config{
		Listener: Listener{
			Mode:         SINGLE_LISTENER_MODE,
			IP:           ip,
			DefaultPort:  "3456",
			DefaultProto: "http",
		},
		Upstream: Upstream{
			SourceType:   "swan",
			PollInterval: time.Second * 30,
		},
		HttpHandler: HttpHandler{
			FlushInterval:  time.Second * 1,
			ClientIPHeader: "",
			Domain:         "dataman-inc.com",
		},
		HttpProxyServer: HttpProxyServer{
			ReadTimeout:  time.Second * 1,
			WriteTimeout: time.Second * 1,
		},
	}

	return config
}

type Config struct {
	Proxy           Proxy
	Upstream        Upstream
	Listener        Listener
	HttpHandler     HttpHandler
	HttpProxyServer HttpProxyServer
}

type Proxy struct {
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

type CertSource struct {
	Name         string
	Type         string
	CertPath     string
	KeyPath      string
	ClientCAPath string
	CAUpgradeCN  string
	Refresh      time.Duration
	Header       http.Header
}

type Upstream struct {
	SourceType   string
	PollInterval time.Duration
}

type Listener struct {
	Mode         string
	IP           string
	DefaultPort  string
	DefaultProto string
}

type HttpHandler struct {
	FlushInterval  time.Duration
	ClientIPHeader string
	Domain         string
}

type HttpProxyServer struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}
