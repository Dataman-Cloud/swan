package main

import (
	"os"

	"time"

	"github.com/Dataman-Cloud/swan-janitor/src/config"
	"github.com/Dataman-Cloud/swan-janitor/src/janitor"
	"github.com/Dataman-Cloud/swan-janitor/src/upstream"

	log "github.com/Sirupsen/logrus"
)

func SetupLogger() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func LoadConfig() config.Config {
	return config.DefaultConfig()
}

func main() {
	janitorConfig := LoadConfig()
	//enable multi_port mode
	//janitorConfig.Listener.Mode = config.MULTIPORT_LISTENER_MODE

	//TuneGolangProcess()
	SetupLogger()

	server := janitor.NewJanitorServer(janitorConfig)
	go server.Init().Run()

	ticker := time.NewTicker(time.Second * 30)
	for {
		<-ticker.C
		log.Debug("sending targetChangeEvent")
		targetChangeEvents := []*upstream.TargetChangeEvent{
			{
				Change:     "add",
				TargetName: "0.nginx0051-01.defaultGroup.dataman-mesos",
				TargetIP:   "192.168.1.162",
				TargetPort: "80",
				//FrontendPort: "8081", //for MULTIPORT_LISTENER_MODE
			},
			{
				Change:     "add",
				TargetName: "1.nginx0051-01.defaultGroup.dataman-mesos",
				TargetIP:   "192.168.1.163",
				TargetPort: "80",
				//FrontendPort: "8081", // for MULTIPORT_LISTENER_MODE
			},
		}

		for _, targetChangeEvent := range targetChangeEvents {
			server.SwanEventChan() <- targetChangeEvent
		}
		time.Sleep(time.Second * 10)
		targetChangeEvents = []*upstream.TargetChangeEvent{
			{
				Change:     "delete",
				TargetName: "0.nginx0051-01.defaultGroup.dataman-mesos",
				TargetIP:   "192.168.1.162",
				TargetPort: "80",
				//FrontendPort: "8081",
			},
			{
				Change:     "delete",
				TargetName: "1.nginx0051-01.defaultGroup.dataman-mesos",
				TargetIP:   "192.168.1.163",
				TargetPort: "80",
				//FrontendPort: "8081",
			},
		}
		for _, targetChangeEvent := range targetChangeEvents {
			server.SwanEventChan() <- targetChangeEvent
		}
		//targetChangeEvent := &upstream.TargetChangeEvent{
		//	Change:       "delete",
		//	TargetName:   "0.nginx0051-01.defaultGroup.dataman-mesos",
		//	TargetIP:     "192.168.1.162",
		//	TargetPort:   "80",
		//	FrontendPort: "8081",
		//}
		//server.SwanEventChan() <- targetChangeEvent
	}
}
