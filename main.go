package main

import (
	"os"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/version"
	"github.com/boltdb/bolt"

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
			Name:  "config-file,c",
			Value: "./config.json",
			Usage: "specify config file path",
		},
		cli.StringFlag{
			Name:   "cluster",
			Usage:  "API Server address <ip:port>",
			EnvVar: "SWAN_CLUSTER",
		},
		cli.StringFlag{
			Name:   "zk",
			Usage:  "Zookeeper URL. eg.zk://host1:port1,host2:port2,.../path",
			EnvVar: "SWAN_ZKURL",
		},
		cli.StringFlag{
			Name:   "log-level,l",
			Usage:  "customize debug level [debug|info|error]",
			EnvVar: "SWAN_LOG_LEVEL",
		},
		cli.IntFlag{
			Name:   "raftid",
			Usage:  "raft node id",
			EnvVar: "SWAN_RAFT_ID",
		},
		cli.StringFlag{
			Name:   "raft-cluster",
			Usage:  "raft cluster peers addr",
			EnvVar: "SWAN_RAFT_CLUSTER",
		},
		cli.StringFlag{
			Name:   "mode",
			Usage:  "Server mode, manager|agent|mixed ",
			EnvVar: "SWAN_MODE",
		},
		cli.StringFlag{
			Name:   "data-dir,d",
			Usage:  "swan data store dir",
			EnvVar: "SWAN_DATA_DIR",
		},
		cli.BoolFlag{
			Name:   "enable-proxy",
			Usage:  "enable proxy or not",
			EnvVar: "SWAN_ENABLE_PROXY",
		},
		cli.BoolFlag{
			Name:   "enable-dns",
			Usage:  "enable dns resolver or not",
			EnvVar: "SWAN_ENABLE_DNS",
		},
		cli.BoolFlag{
			Name:   "no-recover",
			Usage:  "do not recover from previous crush",
			EnvVar: "SWAN_NO_RECOVER",
		},
	}
	app.Action = func(c *cli.Context) error {
		config, err := config.NewConfig(c)
		if err != nil {
			logrus.Errorf("load config failed. Error: %s", err)
			return err
		}

		setupLogger(config.LogLevel)

		os.MkdirAll(config.DataDir, 0700)

		db, err := bolt.Open(config.DataDir+"swan.db", 0600, nil)
		if err != nil {
			logrus.Errorf("Init store engine failed:%s", err)
			return err
		}

		node, err := NewNode(config, db)
		if err != nil {
			logrus.Error("Node initialization failed")
			return err
		}

		if err := node.Start(context.Background()); err != nil {
			logrus.Error("start node failed. Error: %s", err.Error())
			return err
		}

		return nil
	}

	app.Run(os.Args)
}
