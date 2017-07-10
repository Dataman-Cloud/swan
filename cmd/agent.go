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

	agentCmd.Flags = append(agentCmd.Flags, FlagJoinAddrs())
	agentCmd.Flags = append(agentCmd.Flags, FlagGatewayAdvertiseIp())
	agentCmd.Flags = append(agentCmd.Flags, FlagGatewayListenAddr())
	agentCmd.Flags = append(agentCmd.Flags, FlagGatewayTLSListenAddr())
	agentCmd.Flags = append(agentCmd.Flags, FlagGatewayTLSCertFile())
	agentCmd.Flags = append(agentCmd.Flags, FlagGatewayTLSKeyFile())
	agentCmd.Flags = append(agentCmd.Flags, FlagDNSListenAddr())
	agentCmd.Flags = append(agentCmd.Flags, FlagDNSTTL())
	agentCmd.Flags = append(agentCmd.Flags, FlagDNSResolvers())
	agentCmd.Flags = append(agentCmd.Flags, FlagLogLevel())
	agentCmd.Flags = append(agentCmd.Flags, FlagDomain())

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
