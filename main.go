package main

import (
	"github.com/urfave/cli"

	"github.com/Dataman-Cloud/swan/cmd"
	_ "github.com/Dataman-Cloud/swan/debug"
	"github.com/Dataman-Cloud/swan/version"
)

func main() {
	app := cli.NewApp()
	app.Name = "swan"
	app.Description = "A Distributed, Highly Available Mesos Scheduler, Inspired by the design of Google Borg."
	app.Version = version.GetVersion().Version

	app.Commands = []cli.Command{
		cmd.ManagerCmd(),
		cmd.AgentCmd(),
		cmd.VersionCmd(),
	}

	app.RunAndExitOnError()
}
