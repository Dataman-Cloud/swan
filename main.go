package main

import (
	"flag"
	"os"
	"os/user"

	"github.com/Dataman-Cloud/swan/api"
	"github.com/Dataman-Cloud/swan/health"
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/registry/consul"
	"github.com/Dataman-Cloud/swan/scheduler"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	"github.com/gogo/protobuf/proto"
)

var (
	addr       = flag.String("addr", "127.0.0.1:9999", "API Server address <ip:port>")
	master     = flag.String("master", "127.0.0.1:5050", "Master address <ip:port>")
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

	consulClient, err := consul.NewConsul(*consulAddr)
	if err != nil {
		logrus.Errorf("Init consul client failed:%s", err)
		return
	}

	frameworkId, err := consulClient.FetchFrameworkID("swan/frameworkid")
	if err != nil {
		logrus.Errorf("Fetch framework id from consul failed: %s", err)
		return
	}

	if frameworkId != "" {
		fw.Id = &mesos.FrameworkID{
			Value: proto.String(frameworkId),
		}
	}

	reschedQueue := make(chan types.ReschedulerMsg, 1000)

	healthChecker := health.NewHealthChecker(consulClient, reschedQueue)
	healthChecker.Init()

	sched := scheduler.New(*master, fw, consulClient, *clusterId, healthChecker, reschedQueue)

	srv := api.NewServer(sched)
	go func() {
		srv.ListenAndServe(*addr)
	}()

	<-sched.Start()
}
