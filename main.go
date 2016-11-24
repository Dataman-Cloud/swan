package main

import (
	"flag"
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
	"golang.org/x/net/context"
)

var (
	addr    string
	masters string
	user    string
	debug   bool
	raftId  int
	cluster string
)

func init() {
	flag.StringVar(&addr, "addr", "0.0.0.0:9999", "API Server address <ip:port>")
	flag.StringVar(&masters, "masters", "127.0.0.1:5050", "masters address <ip:port>,<ip:port>...")
	flag.StringVar(&user, "user", "root", "mesos user")
	flag.BoolVar(&debug, "debug", false, "log level")
	flag.IntVar(&raftId, "raftid", 1, "raft node id")
	flag.StringVar(&cluster, "cluster", "http://127.0.0.1:2221", "raft cluster peers addr")

	flag.Parse()
}

func setupLogger() {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.SetOutput(os.Stderr)

	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
}

func main() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	fw := &mesos.FrameworkInfo{
		User:            proto.String(user),
		Name:            proto.String("swan"),
		Hostname:        proto.String(hostname),
		FailoverTimeout: proto.Float64(60 * 60 * 24 * 7),
	}

	setupLogger()

	store, err := NewBoltStore(".bolt.db")
	if err != nil {
		logrus.Errorf("Init store engine failed:%s", err)
		return
	}

	raftNode := raft.NewNode(raftId, strings.Split(cluster, ","), store)

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

	masters := []string{masters}
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

	srv := api.NewServer(addr)

	routers := []router.Router{
		application.NewRouter(backend),
	}

	srv.InitRouter(routers...)

	go func() {
		for {
			err := <-ns.StartAsDnsProxy()
			if err != nil {
				logrus.Errorf("start dns proxy got err %s", err)
			}
		}
	}()

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
