package main

import (
	"fmt"
	"os"

	"github.com/Dataman-Cloud/swan/version"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

func setupLogger(c *cli.Context) {
	level, err := logrus.ParseLevel(c.String("log-level"))
	if err != nil {
		level = logrus.DebugLevel
	}

	logrus.SetLevel(level)
	logrus.SetOutput(os.Stdout)

	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s version=%s commit=%s, buildTime=%s\n", c.App.Name, c.App.Version, version.Commit, version.BuildTime)
	}

	app := cli.NewApp()
	app.Name = "swan"
	app.Usage = "next generation mesos framework"
	app.Version = version.Version
	app.Commands = Commands

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "log-level", Value: "debug", Usage: "default log level"},
	}

	app.Before = func(c *cli.Context) error {
		setupLogger(c)

		return nil
	}

	app.Run(os.Args)
}
