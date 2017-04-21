package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

var gc *cli.Context

func PrepareApp() *cli.App {
	app := cli.NewApp()
	app.Name = "swan"
	app.Usage = "swan [ROLE] [COMMAND] [ARG...]"
	app.Description = "A general purpose Mesos framework which facility long running docker application management."

	app.Commands = []cli.Command{}

	app.Commands = append(app.Commands, FakeAgentJoinCmd())
	app.Commands = append(app.Commands, FakeManagerCmd())

	return app
}

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

func FlagJanitorAdvertiseIp() cli.Flag {
	return cli.StringFlag{
		Name:   "janitor-advertise-ip",
		Usage:  "janitor proxy advertise ip",
		EnvVar: "SWAN_JANITOR_ADVERTISE_IP",
		Value:  "",
	}
}

func FlagJanitorListenAddr() cli.Flag {
	return cli.StringFlag{
		Name:   "janitor-listen-addr",
		Usage:  "janitor proxy listen addr",
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

func FakeAgentJoinCmd() cli.Command {
	agentCmd := cli.Command{
		Name:        "agent",
		Usage:       "[COMMAND] [ARG...]",
		Description: "run swan agent command",
		Flags:       []cli.Flag{},
		Subcommands: []cli.Command{},
	}

	agentJoinCmd := cli.Command{
		Name:        "join",
		Usage:       "[COMMAND] [ARG...]",
		Description: "start and join a swan agent which contains proxy and DNS server",
		Flags:       []cli.Flag{},
		Action:      ReturnConfig,
	}

	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagListenAddr())
	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagAdvertiseAddr())
	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagJoinAddrs())
	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagDataDir())
	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagJanitorAdvertiseIp())
	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagJanitorListenAddr())
	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagDNSListenAddr())
	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagDNSResolvers())
	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagLogLevel())
	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagDomain())

	agentCmd.Subcommands = append(agentCmd.Subcommands, agentJoinCmd)

	return agentCmd
}

func FakeManagerCmd() cli.Command {
	managerCmd := cli.Command{
		Name:        "manager",
		Usage:       "[COMMAND] [ARG...]",
		Description: "init a manager as new cluster or join to an exiting cluster",
		Subcommands: []cli.Command{},
	}

	managerCmd.Subcommands = append(managerCmd.Subcommands, ManagerJoinCmd())
	managerCmd.Subcommands = append(managerCmd.Subcommands, FakeManagerInitCmd())

	return managerCmd
}

func ManagerJoinCmd() cli.Command {
	managerJoinCmd := cli.Command{
		Name:        "join",
		Usage:       "join [ARG...]",
		Description: "start a manager and join to an exitsing swan cluster",
		Flags:       []cli.Flag{},
		Action:      ReturnConfig,
	}

	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagListenAddr())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagAdvertiseAddr())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagRaftListenAddr())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagRaftAdvertiseAddr())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagJoinAddrs())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagZkPath())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagLogLevel())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagDataDir())

	return managerJoinCmd
}

func FakeManagerInitCmd() cli.Command {
	managerInitCmd := cli.Command{
		Name:        "init",
		Usage:       "init [ARG...]",
		Description: "start a manager and init a new swan cluster",
		Flags:       []cli.Flag{},
		Action:      ReturnConfig,
	}

	managerInitCmd.Flags = append(managerInitCmd.Flags, FlagListenAddr())
	managerInitCmd.Flags = append(managerInitCmd.Flags, FlagAdvertiseAddr())
	managerInitCmd.Flags = append(managerInitCmd.Flags, FlagRaftListenAddr())
	managerInitCmd.Flags = append(managerInitCmd.Flags, FlagRaftAdvertiseAddr())
	managerInitCmd.Flags = append(managerInitCmd.Flags, FlagZkPath())
	managerInitCmd.Flags = append(managerInitCmd.Flags, FlagLogLevel())
	managerInitCmd.Flags = append(managerInitCmd.Flags, FlagDataDir())

	return managerInitCmd
}

func FakeManagerJoinCmd() cli.Command {
	managerJoinCmd := cli.Command{
		Name:        "join",
		Usage:       "join [ARG...]",
		Description: "start a manager and join to an exitsing swan cluster",
		Flags:       []cli.Flag{},
		Action:      ReturnConfig,
	}

	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagListenAddr())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagAdvertiseAddr())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagRaftListenAddr())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagRaftAdvertiseAddr())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagJoinAddrs())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagZkPath())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagLogLevel())
	managerJoinCmd.Flags = append(managerJoinCmd.Flags, FlagDataDir())

	return managerJoinCmd
}

func ReturnConfig(c *cli.Context) error {
	gc = c
	return nil
}

func TestNewConfig(t *testing.T) {
	assert.True(t, true)
}

func TestValidateAndFormatConfig(t *testing.T) {
	assert.True(t, true)
}

func TestHostname(t *testing.T) {
	osHostName, err := os.Hostname()
	assert.Nil(t, err)
	assert.Equal(t, osHostName, Hostname())
}

func TestNewAgentConfig(t *testing.T) {
	app := PrepareApp()
	app.Run([]string{"swan", "agent", "join", "--join-addrs=132"})
	agentConfig := NewAgentConfig(gc)
	assert.Equal(t, agentConfig.LogLevel, "info")
	assert.Equal(t, "132", gc.String("join-addrs"))
}

func TestNewManagerConfig(t *testing.T) {
	app := PrepareApp()
	app.Run([]string{"swan", "manager", "init", "--zk-path=132"})
	managerConfig, _ := NewManagerConfig(gc)
	assert.Equal(t, managerConfig.LogLevel, "info")
	assert.Equal(t, "132", gc.String("zk-path"))
}
