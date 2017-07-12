package cmd

import (
	"fmt"

	"github.com/Dataman-Cloud/swan/config"
	"github.com/Dataman-Cloud/swan/manager"

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
	managerCmd.Flags = append(managerCmd.Flags, FlagMesosURL())
	managerCmd.Flags = append(managerCmd.Flags, FlagStoreType())
	managerCmd.Flags = append(managerCmd.Flags, FlagZKURL())
	managerCmd.Flags = append(managerCmd.Flags, FlagEtcdAddrs())
	managerCmd.Flags = append(managerCmd.Flags, FlagLogLevel())
	managerCmd.Flags = append(managerCmd.Flags, FlagStrategy())
	managerCmd.Flags = append(managerCmd.Flags, FlagEnableCORS())
	managerCmd.Flags = append(managerCmd.Flags, FlagReconciliationInterval())
	managerCmd.Flags = append(managerCmd.Flags, FlagReconciliationStep())
	managerCmd.Flags = append(managerCmd.Flags, FlagReconciliationStepDelay())

	return managerCmd
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
