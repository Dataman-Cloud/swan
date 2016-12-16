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

// waitForSignals wait for signals and do some clean up job.
func waitForSignals(unixSock string) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for sig := range signals {
		logrus.Debugf("Received signal %s , clean up...", sig)
		if _, err := os.Stat(unixSock); err == nil {
			logrus.Debugf("Remove %s", unixSock)
			if err := os.Remove(unixSock); err != nil {
				logrus.Errorf("Remove %s failed: %s", unixSock, err.Error())
			}
		}

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
			Name:  "config-file",
			Value: "./config.json",
			Usage: "specify config file path",
		},
		cli.StringFlag{
			Name:  "cluster",
			Usage: "API Server address <ip:port>",
		},
		cli.StringFlag{
			Name:  "sock",
			Usage: "Unix socket for listening",
		},
		cli.StringFlag{
			Name:  "master",
			Value: "127.0.0.1:5050",
			Usage: "master address host1:port1,host2:port2,... or zk://host1:port1,host2:port2,.../path",
		},
		cli.StringFlag{
			Name:  "user",
			Usage: "mesos framework user",
		},
		cli.StringFlag{
			Name:  "log-level",
			Usage: "customize debug level [debug|info|error]",
		},
		cli.IntFlag{
			Name:  "raftid",
			Usage: "raft node id",
		},
		cli.StringFlag{
			Name:  "raft-cluster",
			Usage: "raft cluster peers addr",
		},
		cli.BoolFlag{
			Name:  "enable-dns-proxy",
			Usage: "enable dns proxy or not",
		},
		cli.BoolFlag{
			Name:  "enable-local-healthcheck",
			Usage: "Enable local health check",
		},
		cli.StringFlag{
			Name:  "mode",
			Usage: "Server mode, manager|agent|mixed ",
		},
		cli.StringFlag{
			Name:  "data-dir",
			Usage: "swan data store dir",
		},
		cli.StringFlag{
			Name:  "with-engine",
			Value: "sched",
			Usage: "select engine,framework|sched",
		},
		cli.BoolFlag{
			Name:  "enable-proxy",
			Usage: "enable proxy or not",
		},
		cli.BoolFlag{
			Name:  "no-recover",
			Usage: "do not retry recover from previous crush",
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

		node, _ := NewNode(config, db)
		go func() {
			if err := node.Start(context.Background()); err != nil {
				logrus.Errorf("strart node got error: %s", err.Error())
			}
		}()

		waitForSignals(config.HttpListener.UnixAddr)

		return nil
	}

	app.Run(os.Args)
}
