package main

import (
	"flag"
	"os"
	"os/user"

	"github.com/Dataman-Cloud/swan/api"
	"github.com/Dataman-Cloud/swan/inmemory"
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/scheduler"
	"github.com/Sirupsen/logrus"
	"github.com/gogo/protobuf/proto"
	consulapi "github.com/hashicorp/consul/api"
)

var (
	addr      = flag.String("addr", "127.0.0.1:9999", "API Server address <ip:port>")
	master    = flag.String("master", "127.0.0.1:5050", "Master address <ip:port>")
	mesosUser = flag.String("user", "", "Framework user")
	consul    = flag.String("consul", "127.0.0.1:8500", "Consul address <ip:port>")
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

	cfg := consulapi.Config{
		Address: *consul,
		Scheme:  "http",
	}

	consulClient, err := consulapi.NewClient(&cfg)
	if err != nil {
		logrus.Errorf("Init consul client failed:%s", err)
		return
	}

	fw := &mesos.FrameworkInfo{
		User:            mesosUser,
		Name:            proto.String("swan"),
		Hostname:        proto.String(hostname),
		FailoverTimeout: proto.Float64(60 * 60 * 24 * 7),
	}

	kv, _, err := consulClient.KV().Get("swan/frameworkid", nil)
	if err != nil {
		logrus.Errorf("Fetch framework id from consul failed: %s", err)
		return
	}

	if kv != nil {
		frameworkId := string(kv.Value[:])
		fw.Id = &mesos.FrameworkID{
			Value: proto.String(frameworkId),
		}
	}

	sched := scheduler.New(*master, fw, inmemory.New(), consulClient)

	srv := api.NewServer(sched)
	go func() {
		srv.ListenAndServe(*addr)
	}()

	<-sched.Start()
}
