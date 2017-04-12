package cmd

import (
	"github.com/urfave/cli"
)

func FlagListenAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "listen-addr",
		Usage:  "listener address for agent",
		EnvVar: "SWAN_LISTEN_ADDR",
		Value:  "0.0.0.0:9999",
	}
}

func FlagAdvertiseAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "advertise-addr",
		Usage:  "advertise address for agent, default is the listen-addr",
		EnvVar: "SWAN_ADVERTISE_ADDR",
		Value:  "",
	}
}

func FlagRaftListenAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "raft-listen-addr",
		Usage:  "swan raft serverlistener address",
		EnvVar: "SWAN_RAFT_LISTEN_ADDR",
		Value:  "http://0.0.0.0:2111",
	}
}

func FlagRaftAdvertiseAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "raft-advertise-addr",
		Usage:  "swan raft advertise address, default is the raft-listen-addr",
		EnvVar: "SWAN_RAFT_ADVERTISE_ADDR",
		Value:  "",
	}
}

func FlagJoinAddrs() cli.Flag {
	return cli.StringFlag{
		Name:   "join-addrs",
		Usage:  "the addrs new node join to. Splited by ','",
		EnvVar: "SWAN_JOIN_ADDRS",
		Value:  "0.0.0.0:9999",
	}
}

func FlagGossipListenAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "gossip-listen-addr",
		Usage:  "swan gossip node listen address",
		EnvVar: "SWAN_GOSSIP_LISTEN_ADDR",
		Value:  "0.0.0.0:5000",
	}
}

func FlagGossipJoinAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "gossip-join-addr",
		Usage:  "swan gossip node join address",
		EnvVar: "SWAN_GOSSIP_JOIN_ADDR",
	}
}

func FlagJanitorAdvertiseIp() cli.Flag {
	return cli.StringFlag{
		Name:   "janitor-advertise-ip",
		Usage:  "janitor gateway advertise ip",
		EnvVar: "SWAN_JANITOR_ADVERTISE_IP",
		Value:  "",
	}
}

func FlagJanitorListenAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "janitor-listen-addr",
		Usage:  "janitor gateway listen addr",
		Value:  "0.0.0.0:80",
		EnvVar: "SWAN_JANITOR_LISTEN_ADDR",
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

func FlagZkPath() cli.Flag {
	return cli.StringFlag{
		Name:   "zk-path",
		Usage:  "zookeeper mesos paths. eg. zk://host1:port1,host2:port2,.../path",
		EnvVar: "SWAN_MESOS_ZKPATH",
		Value:  "localhost:2181/mesos",
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

func FlagDataDir() cli.Flag {
	return cli.StringFlag{
		Name:   "data-dir,d",
		Usage:  "swan data store dir",
		EnvVar: "SWAN_DATA_DIR",
		Value:  "./data",
	}
}

func FlagDomain() cli.Flag {
	return cli.StringFlag{
		Name:   "domain",
		Usage:  "domain which resolve to proxies. eg. swan.com, which make any task can be access from path likes 0.appname.username.cluster.swan.com",
		EnvVar: "SWAN_DOMAIN",
		Value:  "swan.com",
	}
}
