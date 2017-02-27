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
	app.Usage = "A general purpose Mesos framework which facility long running docker application management."
	app.Version = version.Version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "listen-addr",
			Usage:  "listener address for agent",
			EnvVar: "SWAN_LISTEN_ADDR",
			Value:  "0.0.0.0:9999",
		},
		cli.StringFlag{
			Name:   "advertise-addr",
			Usage:  "advertise address for agent, default is the listen-addr",
			EnvVar: "SWAN_ADVERTISE_ADDR",
			Value:  "",
		},
		cli.StringFlag{
			Name:   "raft-listen-addr",
			Usage:  "swan raft serverlistener address",
			EnvVar: "SWAN_RAFT_LISTEN_ADDR",
			Value:  "http://0.0.0.0:2111",
		},
		cli.StringFlag{
			Name:   "raft-advertise-addr",
			Usage:  "swan raft advertise address, default is the raft-listen-addr",
			EnvVar: "SWAN_RAFT_ADVERTISE_ADDR",
			Value:  "",
		},
		cli.StringFlag{
			Name:   "join-addrs",
			Usage:  "the addrs new node join to. Splited by ','",
			EnvVar: "SWAN_JOIN_ADDRS",
			Value:  "0.0.0.0:9999",
		},
		cli.StringFlag{
			Name:   "janitor-advertise-ip",
			Usage:  "janitor proxy advertise ip",
			EnvVar: "SWAN_JANITOR_ADVERTISE_IP",
			Value:  "",
		},

		cli.StringFlag{
			Name:   "zk-path",
			Usage:  "zookeeper mesos paths. eg. zk://host1:port1,host2:port2,.../path",
			EnvVar: "SWAN_MESOS_ZKPATH",
			Value:  "localhost:2181/mesos",
		},
		cli.StringFlag{
			Name:   "log-level,l",
			Usage:  "customize log level [debug|info|error]",
			EnvVar: "SWAN_LOG_LEVEL",
			Value:  "info",
		},
		cli.StringFlag{
			Name:   "mode",
			Usage:  "server mode, manager|agent",
			EnvVar: "SWAN_MODE",
			Value:  "mixed",
		},
		cli.StringFlag{
			Name:   "data-dir,d",
			Usage:  "swan data store dir",
			EnvVar: "SWAN_DATA_DIR",
			Value:  "./data",
		},
		cli.StringFlag{
			Name:   "domain",
			Usage:  "domain which resolve to proxies. eg. swan.com, which make any task can be access from path likes 0.appname.username.cluster.swan.com",
			EnvVar: "SWAN_DOMAIN",
			Value:  "swan.com",
		},
	}
	app.Action = func(c *cli.Context) error {
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
