package main

import (
	"flag"
	"os"
	"time"

	janitor "github.com/Dataman-Cloud/swan-janitor/src"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

func SetupLogger() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

var listenaddr = flag.String("listenaddr", "0.0.0.0:80", "")
var domain = flag.String("domain", "lvh.me", "")

func main() {
	flag.Parse()
	janitorConfig := janitor.DefaultConfig()
	janitorConfig.ListenAddr = *listenaddr
	janitorConfig.Domain = *domain
	SetupLogger()

	server := janitor.NewJanitorServer(janitorConfig)

	go func() {
		staredCh := make(chan bool)
		err := server.Start(context.Background(), staredCh)
		if err != nil {
			log.Errorf("server start go error: %v", err)
			os.Exit(1)
		}
	}()

	ticker := time.NewTicker(time.Second * 10)
	for {
		<-ticker.C
		log.Debug("sending targetChangeEvent")
		time.Sleep(time.Second * 10)
		targetChangeEvents := []*janitor.TargetChangeEvent{
			{
				Change:   "add",
				AppID:    "nginx0051-xcm-datamanmesos",
				TaskID:   "0-nginx0051-xcm-datamanmesos",
				TaskIP:   "192.168.1.162",
				PortName: "web",
				TaskPort: 80,
			},
			{
				Change:   "add",
				AppID:    "nginx0051-xcm-datamanmesos",
				TaskID:   "1-nginx0051-xcm-datamanmesos",
				TaskIP:   "192.168.1.162",
				PortName: "web1",
				TaskPort: 80,
			},
		}

		for _, targetChangeEvent := range targetChangeEvents {
			server.EventChan <- targetChangeEvent
		}
		time.Sleep(time.Second * 10)
		targetChangeEvents = []*janitor.TargetChangeEvent{
			{
				Change:   "del",
				AppID:    "nginx0051-xcm-datamanmesos",
				TaskID:   "0-nginx0051-xcm-datamanmesos",
				TaskIP:   "192.168.1.162",
				PortName: "web",
				TaskPort: 80,
			},
			{
				Change:   "del",
				AppID:    "nginx0051-xcm-datamanmesos",
				TaskID:   "1-nginx0051-xcm-datamanmesos",
				TaskIP:   "192.168.1.162",
				PortName: "web1",
				TaskPort: 80,
			},
		}
		for _, targetChangeEvent := range targetChangeEvents {
			server.EventChan <- targetChangeEvent
		}
	}
}
