package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/Dataman-Cloud/swan/agent"
	"github.com/Dataman-Cloud/swan/config"
)

func AgentCmd() cli.Command {
	agentCmd := cli.Command{
		Name:        "agent",
		Usage:       "start agent listen for events from leader manager",
		Description: "run swan agent command",
		Flags:       []cli.Flag{},
		Subcommands: []cli.Command{},
		Action:      JoinAndStartAgent,
	}

	agentCmd.Flags = []cli.Flag{
		FlagListenAddr(),
		FlagJoinAddrs(),
		FlagGatewayAdvertiseIp(),
		FlagGatewayListenAddr(),
		FlagGatewayTLSListenAddr(),
		FlagGatewayTLSCertFile(),
		FlagGatewayTLSKeyFile(),
		FlagDNSListenAddr(),
		FlagDNSTTL(),
		FlagDNSResolvers(),
		FlagLogLevel(),
		FlagDomain(),
		FlagIPAMStoreType(),
		FlagIPAMEtcdAddrs(),
		FlagIPAMZKAddrs(),
	}

	return agentCmd
}

func JoinAndStartAgent(c *cli.Context) error {
	conf, err := config.NewAgentConfig(c)
	if err != nil {
		return fmt.Errorf("parse config error: %v", err)
	}

	setupLogger(conf.LogLevel)

	agent := agent.New(conf)
	return agent.StartAndJoin()
}
