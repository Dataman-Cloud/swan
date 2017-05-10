package cmd

import (
	"fmt"
	"os"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

func ManagerCmd() cli.Command {
	managerCmd := cli.Command{
		Name:        "manager",
		Usage:       "start a manager instance",
		Description: "start a swan manager",
		Action:      StartManager,
	}

	managerCmd.Flags = append(managerCmd.Flags, FlagListenAddr())
	managerCmd.Flags = append(managerCmd.Flags, FlagZKURL())
	managerCmd.Flags = append(managerCmd.Flags, FlagMesosURL())
	managerCmd.Flags = append(managerCmd.Flags, FlagLogLevel())

	return managerCmd
}

func StartManager(c *cli.Context) error {
	conf, err := config.NewManagerConfig(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERR] parse config got error: %s\n", err.Error())
		os.Exit(1)
	}

	setupLogger(conf.LogLevel)

	managerNode, err := manager.New(conf)
	if err != nil {
		logrus.Error("Node initialization failed")
		return err
	}

	if err := managerNode.InitAndStart(context.TODO()); err != nil {
		logrus.Errorf("start node failed. Error: %s", err.Error())
		return err
	}

	return nil
}
