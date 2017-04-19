package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/net/context"

	"github.com/Dataman-Cloud/swan/src/agent"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager"
	"github.com/Dataman-Cloud/swan/src/utils"
	"github.com/Dataman-Cloud/swan/src/version"
)

const NodeIDFileName = "ID"

func AgentJoinCmd() cli.Command {
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
		Description: "start and join a swan agent which contains Gateway and DNS server",
		Flags:       []cli.Flag{},
		Action:      JoinAndStartAgent,
	}

	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagListenAddr())
	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagAdvertiseAddr())
	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagJoinAddrs())
	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagGossipJoinAddr())
	agentJoinCmd.Flags = append(agentJoinCmd.Flags, FlagGossipListenAddr())
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

func JoinAndStartAgent(c *cli.Context) error {
	conf := config.NewAgentConfig(c)

	IDFilePath := path.Join(conf.DataDir, NodeIDFileName)
	ID, err := utils.LoadNodeID(IDFilePath)
	if err != nil {
		return err
	}

	setupLogger(conf.LogLevel)

	agent, err := agent.New(ID, conf)
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

func ManagerCmd() cli.Command {
	managerCmd := cli.Command{
		Name:        "manager",
		Usage:       "[COMMAND] [ARG...]",
		Description: "init a manager as new cluster or join to an exiting cluster",
		Subcommands: []cli.Command{},
	}

	managerCmd.Subcommands = append(managerCmd.Subcommands, ManagerJoinCmd())
	managerCmd.Subcommands = append(managerCmd.Subcommands, ManagerInitCmd())

	return managerCmd
}

func ManagerJoinCmd() cli.Command {
	managerJoinCmd := cli.Command{
		Name:        "join",
		Usage:       "join [ARG...]",
		Description: "start a manager and join to an exitsing swan cluster",
		Flags:       []cli.Flag{},
		Action:      JoinAndStartManager,
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

func ManagerInitCmd() cli.Command {
	managerInitCmd := cli.Command{
		Name:        "init",
		Usage:       "init [ARG...]",
		Description: "start a manager and init a new swan cluster",
		Flags:       []cli.Flag{},
		Action:      JoinAndStartManager,
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

func VersionCmd() cli.Command {
	return cli.Command{
		Name:        "version",
		Usage:       "[COMMAND] [ARG...]",
		Description: "show version",
		Action: func(c *cli.Context) error {
			return version.TextFormatTo(os.Stdout)
		},
	}
}

func JoinAndStartManager(c *cli.Context) error {
	conf := config.NewManagerConfig(c)
	setupLogger(conf.LogLevel)

	managerNode, err := manager.New(conf)
	if err != nil {
		logrus.Error("Node initialization failed")
		return err
	}

	fmt.Println("xxxxxxxxxxxxxxxx")

	if err := managerNode.InitAndStart(context.TODO()); err != nil {
		logrus.Errorf("start node failed. Error: %s", err.Error())
		return err
	}

	return nil
}

//func StartManager(c *cli.Context) error {
//conf := config.NewManagerConfig(c)

//IDFilePath := path.Join(conf.DataDir, NodeIDFileName)
//ID, err := utils.LoadNodeID(IDFilePath)
//if err != nil {
//return err
//}

//setupLogger(conf.LogLevel)

//managerNode, err := manager.New(ID, conf)
//if err != nil {
//logrus.Error("Node initialization failed")
//return err
//}

//if err := managerNode.InitAndStart(context.TODO()); err != nil {
//logrus.Errorf("start node failed. Error: %s", err.Error())
//return err
//}

//return nil
//}

func setupLogger(logLevel string) {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.DebugLevel
	}
	logrus.SetLevel(level)

	logrus.SetOutput(os.Stderr)

	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
}
