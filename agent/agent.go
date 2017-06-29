package agent

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"

	"github.com/Dataman-Cloud/swan/agent/janitor"
	"github.com/Dataman-Cloud/swan/agent/janitor/upstream"
	"github.com/Dataman-Cloud/swan/agent/nameserver"
	"github.com/Dataman-Cloud/swan/config"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"
)

type Agent struct {
	config   config.AgentConfig
	resolver *nameserver.Resolver
	janitor  *janitor.JanitorServer
	eventCh  chan []byte
}

func New(cfg config.AgentConfig) *Agent {
	agent := &Agent{
		config:   cfg,
		resolver: nameserver.NewResolver(&cfg.DNS),
		janitor:  janitor.NewJanitorServer(&cfg.Janitor),
		eventCh:  make(chan []byte, 1024),
	}
	return agent
}

func (agent *Agent) StartAndJoin() error {
	errCh := make(chan error)

	go func() {
		err := agent.resolver.Start()
		if err != nil {
			errCh <- err
		}
		logrus.Warnln("resolver quit, error:", err)
	}()

	go func() {
		err := agent.janitor.Start()
		if err != nil {
			errCh <- err
		}
		logrus.Warnln("janitor quit, error:", err)
	}()

	go func() {
		err := agent.apiServe()
		if err != nil {
			errCh <- err
		}
		logrus.Warnln("api server quit, error:", err)
	}()

	go agent.watchEvents()
	go agent.dispatchEvents()

	return <-errCh
}

func (agent *Agent) apiServe() error {
	engine := gin.Default()

	engine.GET("/", func(c *gin.Context) { c.String(200, "OK") })
	engine.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })
	agent.janitor.ApiServe(engine.Group("/proxy"))
	agent.resolver.ApiServe(engine.Group("/dns"))

	return engine.Run(agent.config.ListenAddr)
}

func (agent *Agent) watchEvents() {
	var (
		delayMin = time.Millisecond * 500 // min retry delay 0.5s
		delayMax = time.Second * 60       // max retry delay 60s
		delay    = delayMin               // retry delay
	)

	for {
		managerAddr, err := agent.detectManagerAddr()
		if err != nil {
			logrus.Errorln("agent Join error:", err)
			delay *= 2
			if delay > delayMax {
				delay = delayMax // reset to max
			}
			logrus.Printf("agent Rejoin in %s ...", delay)
			time.Sleep(delay)
			continue
		}

		logrus.Printf("agent Join to manager %s succeed, ready for events ...", managerAddr)
		delay = delayMin // reset to min

		err = agent.watchManagerEvents(managerAddr)
		if err != nil {
			logrus.Errorf("agent watch on manager events error: %v, retry ...", err)
			time.Sleep(time.Second)
		}
	}
}

func (agent *Agent) dispatchEvents() {
	// send local proxy dns record to dns
	var (
		proxyIP = agent.config.Janitor.AdvertiseIP
		ev      = nameserver.BuildRecordEvent("add", "local_proxy", "PROXY", proxyIP, "80", 0, true)
	)
	agent.resolver.EmitChange(ev)

	for evBody := range agent.eventCh {
		var taskEv types.TaskEvent
		if err := json.Unmarshal(evBody, &taskEv); err != nil {
			logrus.Errorf("agent unmarshal task event error: %v", err)
			continue
		}

		if taskEv.GatewayEnabled {
			ev := genJanitorBackendEvent(&taskEv)
			if ev != nil {
				agent.janitor.EmitEvent(ev)
			}
		}

		if taskEv.Type == types.EventTypeTaskHealthy || taskEv.Type == types.EventTypeTaskUnhealthy {
			ev := genDNSRecordEvent(&taskEv)
			if ev != nil {
				agent.resolver.EmitChange(ev)
			}
		}
	}
}

func (agent *Agent) detectManagerAddr() (string, error) {
	for _, addr := range agent.config.JoinAddrs {
		resp, err := http.Get("http://" + addr + "/ping")
		if err != nil {
			logrus.Warnf("detect swan manager %s error %v", addr, err)
			continue
		}
		resp.Body.Close() // prevent fd leak

		logrus.Infof("detect swan manager %s succeed", addr)
		return addr, nil
	}

	return "", errors.New("all of swan manager unavailable")
}

func (agent *Agent) watchManagerEvents(managerAddr string) error {
	eventsDoesMatter := []string{
		types.EventTypeTaskUnhealthy,
		types.EventTypeTaskHealthy,
		types.EventTypeTaskWeightChange,
	}

	eventsPath := fmt.Sprintf("http://%s/v1/events?catchUp=true", managerAddr)
	resp, err := http.Get(eventsPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var (
		ssePrefixData  = "data:"
		ssePrefixEvent = "event:"
		prefixDataLen  = len(ssePrefixData)
		prefixEventLen = len(ssePrefixEvent)
		reader         = bufio.NewReader(resp.Body)
	)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		// skip blank line
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, ssePrefixEvent) {
			eventType := strings.TrimSpace(line[prefixEventLen:])
			if !utils.SliceContains(eventsDoesMatter, eventType) {
				continue
			}

			// read next line of stream
			line, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			// if line is not data section
			if !strings.HasPrefix(line, ssePrefixData) {
				continue
			}

			agent.eventCh <- []byte(line[prefixDataLen:])
		}
	}
}

func genDNSRecordEvent(taskEv *types.TaskEvent) *nameserver.RecordEvent {
	var (
		act    string
		aid    = taskEv.AppID
		tid    = taskEv.TaskID
		ip     = taskEv.IP
		port   = fmt.Sprintf("%d", taskEv.Port)
		weight = taskEv.Weight
	)
	switch taskEv.Type {
	case types.EventTypeTaskHealthy:
		act = "add"
	case types.EventTypeTaskUnhealthy:
		act = "del"
	default:
		return nil
	}

	return nameserver.BuildRecordEvent(act, tid, aid, ip, port, weight, false)
}

func genJanitorBackendEvent(taskEv *types.TaskEvent) *upstream.BackendEvent {
	var (
		act string

		// upstream
		ups    = taskEv.AppID
		alias  = "" // TODO
		listen = "" // TODO

		// backend
		backend = taskEv.TaskID
		ip      = taskEv.IP
		port    = taskEv.Port
		weight  = taskEv.Weight
		version = taskEv.VersionID
	)

	switch taskEv.Type {
	case types.EventTypeTaskHealthy:
		act = "add"
	case types.EventTypeTaskUnhealthy:
		act = "del"
	case types.EventTypeTaskWeightChange:
		act = "change"
	default:
		return nil
	}

	return upstream.BuildBackendEvent(act, ups, alias, listen, backend, ip, version, port, weight)
}
