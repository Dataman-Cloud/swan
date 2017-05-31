package cmd

import (
	"fmt"
	"os"

	"github.com/Dataman-Cloud/swan/config"
	"github.com/Dataman-Cloud/swan/manager"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
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
	managerCmd.Flags = append(managerCmd.Flags, FlagStrategy())
	managerCmd.Flags = append(managerCmd.Flags, FlagEnableCORS())

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
		logrus.Error("Manager initialization failed")
		return err
	}

	if err := managerNode.Start(); err != nil {
		logrus.Errorf("start manager failed. Error: %s", err.Error())
		return err
	}

	return nil
}
