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
			Name:   "cluster-addrs",
			Usage:  "address api server listen on, eg. 192.168.1.1:9999,192.168.1.2:9999",
			EnvVar: "SWAN_CLUSTER_ADDRS",
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
		cli.IntFlag{
			Name:   "raftid",
			Usage:  "unique raft node id within the cluster. should be 1,2 or 3 within a 3-managers cluster",
			EnvVar: "SWAN_RAFT_ID",
		},
		cli.StringFlag{
			Name:   "raft-cluster",
			Usage:  "raft cluster peers. eg. 192.168.1.1:1211,192.168.1.2.1211",
			EnvVar: "SWAN_RAFT_CLUSTER",
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
			Usage:  "domain which resolve to proxies. eg. access a slot by 0.appname.runas.clustername.domain",
			EnvVar: "SWAN_DOMAIN",
		},
		cli.StringFlag{
			Name:   "listen-addr",
			Usage:  "listener address for agent",
			EnvVar: "SWAN_AGENT_LISTEN_ADDR",
		},

		cli.StringFlag{
			Name:   "advertise-addr",
			Usage:  "advertise address for agent",
			EnvVar: "SWAN_AGENT_ADVERTISE_ADDR",
		},

		cli.StringFlag{
			Name:   "janitor-advertise-ip",
			Usage:  "janitor proxy advertise ip",
			EnvVar: "SWAN_JANITOR_ADVERTISE_IP",
		},

		//cli.StringFlag{
		//Name:   "janitor-listen-ip",
		//Usage:  "janitor proxy listener ip",
		//EnvVar: "SWAN_JANITOR_LISTEN_IP",
		//},

		//cli.StringFlag{
		//Name:   "dns-listen-addr",
		//Usage:  "dns proxy listener address",
		//EnvVar: "SWAN_DNS_LISTEN_ADDR",
		//},
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
			logrus.Errorf("start node failed. Error: %s", err.Error())
			return err
		}

		return nil
	}

	app.Run(os.Args)
}
