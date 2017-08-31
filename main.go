package main

import (
	"os"
	"strconv"
	"time"

	"github.com/urfave/cli"

	"github.com/Dataman-Cloud/swan/cmd"
	_ "github.com/Dataman-Cloud/swan/debug"
	"github.com/Dataman-Cloud/swan/version"
)

func main() {
	app := cli.NewApp()
	app.Name = "swan"
	app.Usage = "A Distributed, Highly Available Mesos Scheduler"
	app.Description = "A Distributed, Highly Available Mesos Scheduler, Inspired by the design of Google Borg."
	app.Version = version.GetVersion().Version

	app.Commands = []cli.Command{
		cmd.ManagerCmd(),
		cmd.AgentCmd(),
		cmd.VersionCmd(),
	}

	if delay := os.Getenv("SWAN_START_DELAY"); delay != "" {
		if n, _ := strconv.Atoi(delay); n > 0 {
			time.Sleep(time.Second * time.Duration(n))
		}
	}

	app.RunAndExitOnError()
}
