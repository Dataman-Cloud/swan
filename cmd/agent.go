package cmd

import (
	"errors"
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
		Subcommands: []cli.Command{
			AgentIPAMIPPoolCmd(),
		},
		Action: JoinAndStartAgent,
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

func AgentIPAMIPPoolCmd() cli.Command {
	return cli.Command{
		Name:        "ipam",
		Usage:       "docker IPAM ip pool management",
		Description: "docker IPAM ip pool management",
		Flags: []cli.Flag{
			FlagIPAMIPStart(),
			FlagIPAMIPEnd(),
		},
		Action: IPAMSetIPPool,
	}
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

func IPAMSetIPPool(c *cli.Context) error {
	var (
		ipstart = c.String("ip-start")
		ipend   = c.String("ip-end")
	)

	if ipstart == "" || ipend == "" {
		return errors.New("parameter [ip-start] & [ip-end] required")
	}

	conf, err := config.NewAgentConfig(c)
	if err != nil {
		return fmt.Errorf("parse config error: %v", err)
	}

	agent := agent.New(conf)
	return agent.IPAMSetIPPool(ipstart, ipend)
}
