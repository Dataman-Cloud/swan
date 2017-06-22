package cmd

import (
	"github.com/urfave/cli"
)

func FlagListenAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "listen-addr",
		Usage:  "http listener address",
		EnvVar: "SWAN_LISTEN_ADDR",
		Value:  "0.0.0.0:9999",
	}
}

func FlagAdvertiseAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "advertise-addr",
		Usage:  "http advertised address, default is the listen-addr, used when in docker env",
		EnvVar: "SWAN_ADVERTISE_ADDR",
		Value:  "",
	}
}

func FlagJoinAddrs() cli.Flag {
	return cli.StringFlag{
		Name:   "join-addrs",
		Usage:  "the manager addresses that supposed to join in, splited by ','",
		EnvVar: "SWAN_JOIN_ADDRS",
		Value:  "0.0.0.0:9999",
	}
}

func FlagGatewayAdvertiseIp() cli.Flag {
	return cli.StringFlag{
		Name:   "gateway-advertise-ip",
		Usage:  "gateway advertise ip",
		EnvVar: "SWAN_GATEWAY_ADVERTISE_IP",
		Value:  "",
	}
}

func FlagGatewayListenAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "gateway-listen-addr",
		Usage:  "gateway listen addr",
		Value:  "0.0.0.0:80",
		EnvVar: "SWAN_GATEWAY_LISTEN_ADDR",
	}
}

func FlagGatewayTLSListenAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "gateway-tls-listen-addr",
		Usage:  "gateway tls listen addr",
		Value:  "",
		EnvVar: "SWAN_GATEWAY_TLS_LISTEN_ADDR",
	}
}

func FlagGatewayTLSCertFile() cli.Flag {
	return cli.StringFlag{
		Name:   "gateway-tls-cert-file",
		Usage:  "gateway tls cert file",
		Value:  "",
		EnvVar: "SWAN_GATEWAY_TLS_CERT_FILE",
	}
}

func FlagGatewayTLSKeyFile() cli.Flag {
	return cli.StringFlag{
		Name:   "gateway-tls-key-file",
		Usage:  "gateway tls key file",
		Value:  "",
		EnvVar: "SWAN_GATEWAY_TLS_KEY_FILE",
	}
}

func FlagDNSListenAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "dns-listen-addr",
		Usage:  "dns listen addr",
		Value:  "0.0.0.0:53",
		EnvVar: "SWAN_DNS_LISTEN_ADDR",
	}
}

func FlagDNSTTL() cli.Flag {
	return cli.IntFlag{
		Name:   "dns-ttl",
		Usage:  "dns records ttl",
		Value:  0,
		EnvVar: "SWAN_DNS_TTL",
	}
}

func FlagDNSResolvers() cli.Flag {
	return cli.StringFlag{
		Name:   "dns-resolvers",
		Usage:  "dns resolvers",
		Value:  "114.114.114.114",
		EnvVar: "SWAN_DNS_RESOLVERS",
	}
}

func FlagMesosURL() cli.Flag {
	return cli.StringFlag{
		Name:   "mesos",
		Usage:  "zookeeper mesos paths. eg. zk://host1:port1,host2:port2,.../path",
		EnvVar: "SWAN_MESOS_URL",
	}
}

func FlagZKURL() cli.Flag {
	return cli.StringFlag{
		Name:   "zk",
		Usage:  "eg. zk://host1:port1,host2:port2,.../swan",
		EnvVar: "SWAN_ZK_URL",
	}
}

func FlagLogLevel() cli.Flag {
	return cli.StringFlag{
		Name:   "log-level,l",
		Usage:  "customize log level [debug|info|error]",
		EnvVar: "SWAN_LOG_LEVEL",
		Value:  "info",
	}
}

func FlagDomain() cli.Flag {
	return cli.StringFlag{
		Name:   "domain",
		Usage:  "domain which resolve to gateway. eg. swan.com, which make any task can be access from path likes 0.appname.username.cluster.gateway.swan.com",
		EnvVar: "SWAN_DOMAIN",
		Value:  "swan.com",
	}
}
