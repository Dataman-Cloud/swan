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

func FlagGossipListenAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "gossip-listen-addr",
		Usage:  "swan gossip node listener address",
		EnvVar: "SWAN_GOSSIP_LISTEN_ADDR",
		Value:  "0.0.0.0:5000",
	}
}

func FlagGossipJoinAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "gossip-join-addr",
		Usage:  "any agent in the cluster",
		EnvVar: "SWAN_GOSSIP_JOIN_ADDR",
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

func FlagDNSListenAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "dns-listen-addr",
		Usage:  "dns listen addr",
		Value:  "0.0.0.0:53",
		EnvVar: "SWAN_DNS_LISTEN_ADDR",
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

func FlagMesosZkPath() cli.Flag {
	return cli.StringFlag{
		Name:   "mesos-zk-path",
		Usage:  "zookeeper mesos paths. eg. zk://host1:port1,host2:port2,.../path",
		EnvVar: "SWAN_MESOS_ZKPATH",
	}
}

func FlagZkPath() cli.Flag {
	return cli.StringFlag{
		Name:   "zk-path",
		Usage:  "eg. zk://host1:port1,host2:port2,.../swan",
		EnvVar: "SWAN_ZKPATH",
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
