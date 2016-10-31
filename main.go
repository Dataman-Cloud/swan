package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/Dataman-Cloud/swan/api"
	"github.com/Dataman-Cloud/swan/backend"
	"github.com/Dataman-Cloud/swan/health"
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/registry/consul"
	"github.com/Dataman-Cloud/swan/scheduler"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/samuel/go-zookeeper/zk"
)

var (
	addr       = flag.String("addr", "127.0.0.1:9999", "API Server address <ip:port>")
	zks        = flag.String("zks", "127.0.0.1:2181", "Zookeeper address <ip:port>")
	mesosUser  = flag.String("user", "", "Framework user")
	consulAddr = flag.String("consul", "127.0.0.1:8500", "Consul address <ip:port>")
	clusterId  = flag.String("cluster_id", "00001", "mesos cluster id")
)

func init() {
	flag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
}

func main() {
	if *mesosUser == "" {
		u, err := user.Current()
		if err != nil {
			logrus.Fatal("Unable to determine user")
		}
		*mesosUser = u.Username
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	fw := &mesos.FrameworkInfo{
		User:            mesosUser,
		Name:            proto.String("swan"),
		Hostname:        proto.String(hostname),
		FailoverTimeout: proto.Float64(60 * 60 * 24 * 7),
	}

	store, err := consul.NewConsul(*consulAddr)
	if err != nil {
		logrus.Errorf("Init store engine failed:%s", err)
		return
	}

	frameworkId, err := store.FetchFrameworkID("swan/frameworkid")
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

	zks := []string{*zks}
	conn, _, err := zk.Connect(zks, 5*time.Second)
	defer conn.Close()
	if err != nil {
		logrus.Error("Couldn't connect to zookeeper")
	}

	children, _, _ := conn.Children("/mesos")
	masterInfo := new(mesos.MasterInfo)
	for _, name := range children {
		data, _, _ := conn.Get("/mesos" + "/" + name)
		err := json.Unmarshal(data, masterInfo)
		if err == nil {
			break
		}
	}

	sched := scheduler.New(
		fmt.Sprintf("%s:%d", masterInfo.GetHostname(), masterInfo.GetPort()),
		fw,
		store,
		*clusterId,
		health.NewHealthCheckManager(store, msgQueue),
		msgQueue,
	)

	backend := backend.NewBackend(sched, store)

	srv := api.NewServer(backend)
	go func() {
		srv.ListenAndServe(*addr)
	}()

	<-sched.Start()
}
