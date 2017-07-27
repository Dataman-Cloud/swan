package cmd

import (
	"fmt"

	"github.com/Dataman-Cloud/swan/config"
	"github.com/Dataman-Cloud/swan/manager"

	"github.com/urfave/cli"
)

func ManagerCmd() cli.Command {
	cmd := cli.Command{
		Name:        "manager",
		Usage:       "start a manager instance",
		Description: "start a swan manager",
		Action:      StartManager,
	}

	cmd.Flags = []cli.Flag{
		FlagListenAddr(),
		FlagMesosURL(),
		FlagStoreType(),
		FlagZKURL(),
		FlagEtcdAddrs(),
		FlagLogLevel(),
		FlagStrategy(),
		FlagEnableCORS(),
		FlagReconciliationInterval(),
		FlagReconciliationStep(),
		FlagReconciliationStepDelay(),
		FlagHeartbeatTimeout(),
		FlagMaxTasksPerOffer(),
	}

	return cmd
}

func StartManager(c *cli.Context) error {
	conf, err := config.NewManagerConfig(c)
	if err != nil {
		return fmt.Errorf("parse config error: %v", err)
	}

	setupLogger(conf.LogLevel)

	managerNode, err := manager.New(conf)
	if err != nil {
		return fmt.Errorf("initilize manager failed: %v", err)
	}

	return managerNode.Start()
}
