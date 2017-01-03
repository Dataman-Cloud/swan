package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

func waitForSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for sig := range signals {
		logrus.Debugf("Received signal %s , clean up...", sig)
		os.Exit(0)
	}
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
			Name:   "mesos-master,m",
			Usage:  "mesos master address host1:port1,host2:port2,... or zk://host1:port1,host2:port2,.../path",
			EnvVar: "SWAN_MESOS_MASTER",
		},
		cli.StringFlag{
			Name:  "log-level,l",
			Usage: "customize debug level [debug|info|error]",
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
			os.Exit(1)
		}

		setupLogger(config.LogLevel)

		db, err := bolt.Open(fmt.Sprintf(config.DataDir+"bolt.db.%d", config.Raft.RaftId), 0600, nil)
		if err != nil {
			logrus.Errorf("Init store engine failed:%s", err)
			return err
		}

		node, err := NewNode(config, db)
		if err != nil {
			logrus.Fatal("Node initialization failed")
		}
		go func() {
			if err := node.Start(context.Background()); err != nil {
				logrus.Errorf("strart node got error: %s", err.Error())
				os.Exit(1)
			}
		}()

		waitForSignals()

		return nil
	}

	app.Run(os.Args)
}
