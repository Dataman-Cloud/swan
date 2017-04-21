package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/Dataman-Cloud/swan/src/cmd"
	_ "github.com/Dataman-Cloud/swan/src/debug"
	"github.com/Dataman-Cloud/swan/src/version"
)

func main() {
	app := cli.NewApp()
	app.Name = "swan"
	app.Usage = "swan [COMMAND] [ARGS]"
	app.Description = "A general purpose Mesos framework which facility long running docker application management."
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
