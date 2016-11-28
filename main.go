package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Dataman-Cloud/swan/manager/raft"
	. "github.com/Dataman-Cloud/swan/store/local"
	"github.com/Dataman-Cloud/swan/util"

	"github.com/Sirupsen/logrus"
	events "github.com/docker/go-events"
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
	app.Version = "0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "addr",
			Value: "0.0.0.0:9999",
			Usage: "API Server address <ip:port>",
		},
		cli.StringFlag{
			Name:  "sock",
			Value: "/var/run/swan.sock",
			Usage: "Unix socket for listening",
		},
		cli.StringFlag{
			Name:  "masters",
			Value: "127.0.0.1:5050",
			Usage: "masters address <ip:port>,<ip:port>...",
		},
		cli.StringFlag{
			Name:  "user",
			Value: "root",
			Usage: "mesos framework user",
		},
		cli.StringFlag{
			Name:  "log-level",
			Value: "info",
			Usage: "customize debug level [debug|info|error]",
		},
		cli.IntFlag{
			Name:  "raftid",
			Value: 1,
			Usage: "raft node id",
		},
		cli.StringFlag{
			Name:  "cluster",
			Value: "http://127.0.0.1.2221",
			Usage: "raft cluster peers addr",
		},
		cli.BoolTFlag{
			Name:  "enable-dns-proxy",
			Usage: "enable dns proxy or not",
		},
		cli.BoolFlag{
			Name:  "enable-local-healthcheck",
			Usage: "Enable local healt check",
		},
		cli.BoolFlag{
			Name:  "standalone",
			Usage: "Run as standalone mode",
		},
		cli.StringFlag{
			Name:  "mode",
			Value: "mixed",
			Usage: "Server mode, manager|agent|mixed ",
		},
	}

	app.Action = func(c *cli.Context) error {
		config, err := util.NewConfig(c)
		if err != nil {
			os.Exit(1)
		}

		setupLogger(config.LogLevel)

		doneCh := make(chan bool)
		node, _ := NewNode(config)
		go func() {
			node.Start(context.Background())
		}()

		Start(config)
		<-doneCh

		return nil
	}

	app.Run(os.Args)
}

func Start(config util.SwanConfig) {
	store, err := NewBoltStore(".bolt.db")
	if err != nil {
		logrus.Errorf("Init store engine failed:%s", err)
		return
	}

	if config.Standalone {
		raftNode := raft.NewNode(config.Raft.RaftId, strings.Split(config.Raft.Cluster, ","), store)

		leadershipCh, cancel := raftNode.SubscribeLeadership()
		defer cancel()

		go handleLeadershipEvents(context.TODO(), leadershipCh)

		ctx := context.Background()
		go func() {
			err := raftNode.StartRaft(ctx)
			if err != nil {
				log.Fatal(err)
			}
		}()

		if err := raftNode.WaitForLeader(ctx); err != nil {
			panic(err)
		}
	}

}

func handleLeadershipEvents(ctx context.Context, leadershipCh chan events.Event) {
	for {
		select {
		case leadershipEvent := <-leadershipCh:
			// TODO lock it and if manager stop return
			newState := leadershipEvent.(raft.LeadershipState)

			if newState == raft.IsLeader {
				fmt.Println("Now i am a leader !!!!!")
			} else if newState == raft.IsFollower {
				fmt.Println("Now i am a follower !!!!!")
			}
		case <-ctx.Done():
			return
		}
	}
}
