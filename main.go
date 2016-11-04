package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/Dataman-Cloud/swan/api"
	"github.com/Dataman-Cloud/swan/backend"
	"github.com/Dataman-Cloud/swan/health"
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/scheduler"
	"github.com/Dataman-Cloud/swan/store/boltdb"
	"github.com/Dataman-Cloud/swan/types"

	"github.com/Sirupsen/logrus"
	"github.com/andygrunwald/megos"
	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
)

var (
	addr       string
	masters    string
	user       string
	consulAddr string
	debug      bool
)

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1:9999", "API Server address <ip:port>")
	flag.StringVar(&masters, "masters", "127.0.0.1:5050", "masters address <ip:port>,<ip:port>...")
	flag.StringVar(&user, "user", "root", "mesos user")
	flag.StringVar(&consulAddr, "consul", "127.0.0.1:8500", "Consul address <ip:port>")
	flag.BoolVar(&debug, "debug", false, "log level")

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

	db, err := bolt.Open("bolt.db", 0644, nil)
	if err != nil {
		logrus.Errorf("Init store engine failed:%s", err)
		return
	}
	defer db.Close()

	store := boltdb.NewBoltdbStore(db)

	frameworkId, err := store.GetFrameworkID()
	if err != nil {
		logrus.Errorf("get framework id failed: %s", err)
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

	srv := api.NewServer(backend)
	go func() {
		srv.ListenAndServe(addr)
	}()

	<-sched.Start()
}
