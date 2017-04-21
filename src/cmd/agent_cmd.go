package cmd

import (
	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/net/context"

	"github.com/Dataman-Cloud/swan/src/agent"
	"github.com/Dataman-Cloud/swan/src/config"
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

	agentCmd.Flags = append(agentCmd.Flags, FlagListenAddr())
	agentCmd.Flags = append(agentCmd.Flags, FlagAdvertiseAddr())
	agentCmd.Flags = append(agentCmd.Flags, FlagJoinAddrs())
	agentCmd.Flags = append(agentCmd.Flags, FlagGossipJoinAddr())
	agentCmd.Flags = append(agentCmd.Flags, FlagGossipListenAddr())
	agentCmd.Flags = append(agentCmd.Flags, FlagGatewayAdvertiseIp())
	agentCmd.Flags = append(agentCmd.Flags, FlagGatewayListenAddr())
	agentCmd.Flags = append(agentCmd.Flags, FlagDNSListenAddr())
	agentCmd.Flags = append(agentCmd.Flags, FlagDNSResolvers())
	agentCmd.Flags = append(agentCmd.Flags, FlagLogLevel())
	agentCmd.Flags = append(agentCmd.Flags, FlagDomain())

	return agentCmd
}

func JoinAndStartAgent(c *cli.Context) error {
	conf := config.NewAgentConfig(c)

	setupLogger(conf.LogLevel)

	agent, err := agent.New(conf)
	if err != nil {
		logrus.Error("agent initialization failed")
		return err
	}

	if err := agent.StartAndJoin(context.TODO()); err != nil {
		logrus.Errorf("start node failed. Error: %s", err.Error())
		return err
	}

	return nil
}
