package cmd

import (
	"github.com/urfave/cli"
)

func FlagListenAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "listen",
		Usage:  "http listener address",
		EnvVar: "SWAN_LISTEN_ADDR",
		Value:  "0.0.0.0:9999",
	}
}

func FlagAdvertiseAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "advertise",
		Usage:  "http advertise address",
		EnvVar: "SWAN_ADVERTISE_ADDR",
		Value:  "",
	}
}

func FlagStoreType() cli.Flag {
	return cli.StringFlag{
		Name:   "store-type",
		Usage:  "db store type [etcd|zk]",
		EnvVar: "SWAN_STORE_TYPE",
		Value:  "zk",
	}
}

func FlagEtcdAddrs() cli.Flag {
	return cli.StringFlag{
		Name:   "etcd-addrs",
		Usage:  "etcd cluster address",
		EnvVar: "SWAN_ETCD_ADDRS",
	}
}

func FlagStrategy() cli.Flag {
	return cli.StringFlag{
		Name:   "strategy",
		Usage:  "scheduler strategy",
		EnvVar: "SWAN_SCHEDULER_STRATEGY",
		Value:  "spread",
	}
}

func FlagEnableCORS() cli.Flag {
	return cli.BoolTFlag{
		Name:   "enable-cors",
		Usage:  "enable cross-origin resource sharing",
		EnvVar: "SWAN_ENABLE_CORS",
	}
}

func FlagReconciliationInterval() cli.Flag {
	return cli.Float64Flag{
		Name:   "reconciliation-interval",
		Usage:  "The period, in seconds, between task reconciliation operations.",
		EnvVar: "SWAN_RECONCILIATION_INTERVAL",
		Value:  900,
	}
}

func FlagReconciliationStep() cli.Flag {
	return cli.Int64Flag{
		Name:   "reconciliation-step",
		Usage:  "The number of tasks reconciled each time",
		EnvVar: "SWAN_RECONCILIATION_STEP",
		Value:  100,
	}
}

func FlagReconciliationStepDelay() cli.Flag {
	return cli.Float64Flag{
		Name:   "reconciliation-step-delay",
		Usage:  "The delay, in seconds, for each step of task reconciliation",
		EnvVar: "SWAN_RECONCILIATION_DELAY",
		Value:  15,
	}
}

func FlagHeartbeatTimeout() cli.Flag {
	return cli.Float64Flag{
		Name:   "heartbeat-timeout",
		Usage:  "The timeout, before to reconnect to mesos",
		EnvVar: "SWAN_HEARTBEAT_TIMEOUT",
		Value:  30,
	}
}

func FlagMaxTasksPerOffer() cli.Flag {
	return cli.IntFlag{
		Name:   "max-tasks-per-offer",
		Usage:  "Launch at most this number of tasks per offer",
		EnvVar: "SWAN_MAX_TASKS_PER_OFFER",
		Value:  1,
	}
}

func FlagEnableCapabilityKilling() cli.Flag {
	return cli.StringFlag{
		Name:   "enable-capability-killing",
		Usage:  "To enable TASK_KILLING state in Mesos (0.28 or later)",
		EnvVar: "SWAN_ENABLE_CAPABILITY_KILLING",
		Value:  "false",
	}
}

func FlagEnableCheckPoint() cli.Flag {
	return cli.StringFlag{
		Name:   "enable-checkpoint",
		Usage:  "To enable check point mechanism on mesos",
		EnvVar: "SWAN_ENABLE_CHECK_POINT",
		Value:  "false",
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

// Gateway
//
func FlagGatewayEnabled() cli.Flag {
	return cli.StringFlag{
		Name:   "gateway-enabled",
		Usage:  "proxy gateway enable or not",
		EnvVar: "SWAN_GATEWAY_ENABLED",
		Value:  "true",
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

// Dns
//
func FlagDNSEnabled() cli.Flag {
	return cli.StringFlag{
		Name:   "dns-enabled",
		Usage:  "dns enable or not",
		EnvVar: "SWAN_DNS_ENABLED",
		Value:  "true",
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

// Agent IPAM
//
func FlagIPAMEnabled() cli.Flag {
	return cli.StringFlag{
		Name:   "ipam-enabled",
		Usage:  "ipam enable or not",
		EnvVar: "SWAN_IPAM_ENABLED",
		Value:  "false",
	}
}

func FlagIPAMStoreType() cli.Flag {
	return cli.StringFlag{
		Name:   "ipam-store-type",
		Usage:  "ipam store type [etcd|zk]",
		EnvVar: "SWAN_IPAM_STORE_TYPE",
	}
}

func FlagIPAMEtcdAddrs() cli.Flag {
	return cli.StringFlag{
		Name:   "ipam-etcd-addrs",
		Usage:  "ipam etcd cluster address",
		EnvVar: "SWAN_IPAM_ETCD_ADDRS",
	}
}

func FlagIPAMZKAddrs() cli.Flag {
	return cli.StringFlag{
		Name:   "ipam-zk-addrs",
		Usage:  "ipam zk cluster address",
		EnvVar: "SWAN_IPAM_ZK_ADDRS",
	}
}

// Agent IPAM IPPool
//
func FlagIPAMIPStart() cli.Flag {
	return cli.StringFlag{
		Name:  "ip-start",
		Usage: "ip pool start, mus be CIDR format",
	}
}

func FlagIPAMIPEnd() cli.Flag {
	return cli.StringFlag{
		Name:  "ip-end",
		Usage: "ip pool end, must be CIDR format",
	}
}
