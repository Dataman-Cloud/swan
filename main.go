package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli"
	"golang.org/x/net/context"

	"github.com/Dataman-Cloud/swan/src/cmd"
	_ "github.com/Dataman-Cloud/swan/src/debug"
	"github.com/Dataman-Cloud/swan/src/version"
)

func main() {
	app := cli.NewApp()
	app.Name = "swan"
	app.Usage = "swan [COMMAND] [ARGS]"
	app.Description = "A Distributed, Highly Available Mesos Scheduler, Inspired by the design of Google Borg."
	app.Version = version.GetVersion().Version

	app.Commands = []cli.Command{}

	ctx, cancel := context.WithCancel(context.TODO())

	app.Commands = append(app.Commands, cmd.ManagerCmd(ctx))
	app.Commands = append(app.Commands, cmd.AgentCmd(ctx))
	app.Commands = append(app.Commands, cmd.VersionCmd())

	quit := make(chan error)
	go func() {
		quit <- app.Run(os.Args)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-sig:
			cancel()
			log.Fatalf("Signal (%s) received, exitting\n", s.String())
		case e := <-quit:
			cancel()
			log.Fatalf("Exitting %s", e)
		}
	}
}
