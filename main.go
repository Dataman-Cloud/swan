package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/Dataman-Cloud/swan/cmd"
	_ "github.com/Dataman-Cloud/swan/debug"
	"github.com/Dataman-Cloud/swan/version"
)

func main() {
	app := cli.NewApp()
	app.Name = "swan"
	app.Usage = "swan [COMMAND] [ARGS]"
	app.Description = "A Distributed, Highly Available Mesos Scheduler, Inspired by the design of Google Borg."
	app.Version = version.GetVersion().Version

	app.Commands = []cli.Command{}

	app.Commands = append(app.Commands, cmd.ManagerCmd())
	app.Commands = append(app.Commands, cmd.AgentCmd())
	app.Commands = append(app.Commands, cmd.VersionCmd())

	if err := app.Run(os.Args); err != nil {
		logrus.Errorf("%s", err.Error())
		os.Exit(1)
	}
}
