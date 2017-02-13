package main

import (
	"os"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/node"
	"github.com/Dataman-Cloud/swan/src/version"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

func setupLogger(logLevel string) {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.DebugLevel
	}
	logrus.SetLevel(level)

	logrus.SetOutput(os.Stderr)

	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
}

func main() {
	app := cli.NewApp()
	app.Name = "swan"
	app.Usage = "A general purpose mesos framework"
	app.Version = version.Version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "listen-addr",
			Usage:  "listener address for agent",
			EnvVar: "SWAN_LISTEN_ADDR",
		},
		cli.StringFlag{
			Name:   "advertise-addr",
			Usage:  "advertise address for agent",
			EnvVar: "SWAN_ADVERTISE_ADDR",
		},
		cli.StringFlag{
			Name:   "raft-listen-addr",
			Usage:  "swan raft serverlistener address",
			EnvVar: "SWAN_RAFT_LISTEN_ADDR",
		},
		cli.StringFlag{
			Name:   "raft-advertise-addr",
			Usage:  "swan raft advertise address",
			EnvVar: "SWAN_RAFT_ADVERTISE_ADDR",
		},
		cli.StringFlag{
			Name:   "join-addrs",
			Usage:  "the addrs new node join to. Splited by ','",
			EnvVar: "SWAN_JOIN_ADDRS",
		},
		cli.StringFlag{
			Name:   "janitor-advertise-ip",
			Usage:  "janitor proxy advertise ip",
			EnvVar: "SWAN_JANITOR_ADVERTISE_IP",
		},

		cli.StringFlag{
			Name:   "zk-path",
			Usage:  "zookeeper mesos paths. eg. zk://host1:port1,host2:port2,.../path",
			EnvVar: "SWAN_MESOS_ZKPATH",
		},
		cli.StringFlag{
			Name:   "log-level,l",
			Usage:  "customize debug level [debug|info|error]",
			EnvVar: "SWAN_LOG_LEVEL",
		},
		cli.StringFlag{
			Name:   "mode",
			Usage:  "server mode, manager|agent|mixed ",
			EnvVar: "SWAN_MODE",
		},
		cli.StringFlag{
			Name:   "data-dir,d",
			Usage:  "swan data store dir",
			EnvVar: "SWAN_DATA_DIR",
		},
		cli.StringFlag{
			Name:   "domain",
			Usage:  "domain which resolve to proxies. eg. swan.com, which make any task can be access from path likes 0.appname.username.cluster.swan.com",
			EnvVar: "SWAN_DOMAIN",
		},
	}
	app.Action = func(c *cli.Context) error {
		if c.NumFlags() == 0 {
			return cli.ShowAppHelp(c)
		}

		config, err := config.NewConfig(c)
		if err != nil {
			logrus.Errorf("load config failed. Error: %s", err)
			return err
		}

		setupLogger(config.LogLevel)

		node, err := node.NewNode(config)
		if err != nil {
			logrus.Error("Node initialization failed")
			return err
		}

		if err := node.Start(context.Background()); err != nil {
			logrus.Errorf("start node failed. Error: %s", err.Error())
			return err
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Errorf("%s", err.Error())
		os.Exit(1)
	}
}
