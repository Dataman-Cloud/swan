package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/Dataman-Cloud/swan/api"
	"github.com/Dataman-Cloud/swan/api/router"
	"github.com/Dataman-Cloud/swan/api/router/application"
	"github.com/Dataman-Cloud/swan/backend"
	"github.com/Dataman-Cloud/swan/health"
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/ns"
	"github.com/Dataman-Cloud/swan/raft"
	"github.com/Dataman-Cloud/swan/scheduler"
	. "github.com/Dataman-Cloud/swan/store/local"
	"github.com/Dataman-Cloud/swan/types"

	"github.com/Sirupsen/logrus"
	"github.com/andygrunwald/megos"
	events "github.com/docker/go-events"
	"github.com/golang/protobuf/proto"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

func setupLogger(c *cli.Context) {
	level, err := logrus.ParseLevel(c.String("log-level"))
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
			Value: "debug",
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
	}

	app.Action = func(c *cli.Context) error {
		Start(c)
		return nil
	}

	app.Run(os.Args)
}

func Start(c *cli.Context) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	if c.Bool("enable-dns-proxy") {
		if os.Getuid() == 0 || (len(os.Getenv("SUDO_UID")) > 0) {
			go func() {
				err := <-ns.StartAsDnsProxy()
				if err != nil {
					logrus.Errorf("start dns proxy got err %s", err)
				}
			}()
		} else {
			logrus.Errorf("no permission to run dns proxy")
			os.Exit(1)
		}
	}

	fw := &mesos.FrameworkInfo{
		User:            proto.String(c.String("user")),
		Name:            proto.String("swan"),
		Hostname:        proto.String(hostname),
		FailoverTimeout: proto.Float64(60 * 60 * 24 * 7),
	}

	setupLogger(c)

	store, err := NewBoltStore(".bolt.db")
	if err != nil {
		logrus.Errorf("Init store engine failed:%s", err)
		return
	}

	raftNode := raft.NewNode(c.Int("raftid"), strings.Split(c.String("cluster"), ","), store)

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

	frameworkId, err := store.FetchFrameworkID()
	if err != nil {
		logrus.Errorf("Fetch framework id failed: %s", err)
		return
	}

	if frameworkId != "" {
		fw.Id = &mesos.FrameworkID{
			Value: proto.String(frameworkId),
		}
	}

	msgQueue := make(chan types.ReschedulerMsg, 1)

	masters := []string{c.String("masters")}
	masterUrls := make([]*url.URL, 0)
	for _, master := range masters {
		masterUrl, _ := url.Parse(fmt.Sprintf("http://%s", master))
		masterUrls = append(masterUrls, masterUrl)
	}

	mesos := megos.NewClient(masterUrls, nil)
	state, err := mesos.GetStateFromCluster()
	if err != nil {
		panic(err)
	}

	cluster := state.Cluster
	if cluster == "" {
		cluster = "Unnamed"
	}

	sched := scheduler.NewScheduler(
		state.Leader,
		fw,
		store,
		cluster,
		health.NewHealthCheckManager(store, msgQueue),
		msgQueue,
	)

	backend := backend.NewBackend(sched, store)

	srv := api.NewServer(c.String("addr"), c.String("sock"))

	routers := []router.Router{
		application.NewRouter(backend),
	}

	srv.InitRouter(routers...)

	go func() {
		srv.ListenAndServe()
	}()

	<-sched.Start()
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
