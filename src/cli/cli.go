package main

import (
	"fmt"
	"github.com/Dataman-Cloud/swan/src/cli/command"
	"github.com/urfave/cli"
	"os"
)

func main() {
	swan := cli.NewApp()
	swan.Name = "swancfg"
	swan.Usage = "command-line client for swan"
	swan.Version = "0.1"
	swan.Copyright = "(c) 2016 Dataman Cloud"

	swan.Commands = []cli.Command{
		command.NewRunCommand(),
		command.NewInspectCommand(),
		command.NewListCommand(),
		command.NewDeleteCommand(),
		command.NewScaleUpCommand(),
		command.NewScaleDownCommand(),
		command.NewUpdateCommand(),
		command.NewProceedUpdateCommand(),
		command.NewCancelUpdateCommand(),
	}

	if err := swan.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
