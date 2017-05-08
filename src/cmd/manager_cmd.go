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

func ManagerCmd(ctx context.Context) cli.Command {
	managerCmd := cli.Command{
		Name:        "manager",
		Usage:       "start a manager instance",
		Description: "start a swan manager",
		Action:      StartManager(ctx),
	}

	managerCmd.Flags = append(managerCmd.Flags, FlagListenAddr())
	managerCmd.Flags = append(managerCmd.Flags, FlagZkPath())
	managerCmd.Flags = append(managerCmd.Flags, FlagMesosZkPath())
	managerCmd.Flags = append(managerCmd.Flags, FlagLogLevel())

	return managerCmd
}

func StartManager(ctx context.Context) func(*cli.Context) error {
	return func(c *cli.Context) error {

		conf, err := config.NewManagerConfig(c)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERR] parse config got error: %s\n", err.Error())
			os.Exit(1)
		}

		setupLogger(conf.LogLevel)

		mgr, err := manager.New(conf)
		if err != nil {
			logrus.Error("Node initialization failed")
			return err
		}

		if err := mgr.Start(ctx); err != nil {
			logrus.Errorf("start node failed. Error: %s", err.Error())
			return err
		}

		return nil
	}
}
