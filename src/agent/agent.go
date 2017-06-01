package agent

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/src/config"
	eventbus "github.com/Dataman-Cloud/swan/src/event"
	"github.com/Dataman-Cloud/swan/src/janitor"
	"github.com/Dataman-Cloud/swan/src/nameserver"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/utils"
	"github.com/Dataman-Cloud/swan/src/utils/httpclient"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/serf/serf"
	"golang.org/x/net/context"
)

const REJOIN_BACKOFF = 3 * time.Second
const SSE_DATA_PREFIX = "data:"
const SSE_EVENT_PREFIX = "event:"
const SSE_BLANK_LINE = ""

// Agent struct
type Agent struct {
	Resolver   *nameserver.Resolver
	Janitor    *janitor.JanitorServer
	HTTPServer *HTTPServer
	SerfServer *SerfServer

	Config config.AgentConfig
}

// New agent func
func New(agentConf config.AgentConfig) *Agent {
	agent := &Agent{
		Config:     agentConf,
		Resolver:   nameserver.NewResolver(&agentConf.DNS),
		Janitor:    janitor.NewJanitorServer(&agentConf.Janitor),
		SerfServer: NewSerfServer(agentConf.GossipListenAddr, agentConf.GossipJoinAddr),
	}
	agent.HTTPServer = NewHTTPServer(agentConf.ListenAddr, agent)
	return agent
}

// StartAndJoin func
func (agent *Agent) StartAndJoin() error {
	errCh := make(chan error)

	go func() {
		err := agent.Resolver.Start(agent.Config.DNS.Domain)
		if err != nil {
			errCh <- err
		}
		logrus.Warnln("resolver quit, error:", err)
	}()

	go func() {
		err := agent.Janitor.Start()
		if err != nil {
			errCh <- err
		}
		logrus.Warnln("janitor quit, error:", err)
	}()

	go func() {
		err := agent.SerfServer.Start()
		if err != nil {
			errCh <- err
		}
		logrus.Warnln("serf server quit, error:", err)
	}()

	go func() {
		err := agent.HTTPServer.Start()
		if err != nil {
			errCh <- err
		}
		logrus.Warnln("http server quit, error:", err)
	}()

	go agent.watchEvents()
	go agent.dispatchEvents()

	return <-errCh
}

// watchEvents establish a connection to swan master's stream events endpoint
// and broadcast received events
func (agent *Agent) watchEvents() {
	for {
		leaderAddr, err := agent.detectManagerLeader()
		if err != nil {
			logrus.Errorf("detect manager leader got error: %v, retry ...", err)
			time.Sleep(REJOIN_BACKOFF)
			continue
		}
		logrus.Printf("detected manager addr %s, listening on events ...", leaderAddr)

		err = agent.watchManagerEvents(leaderAddr)
		if err != nil {
			logrus.Errorf("watch manager events got error: %v, retry ...", err)
			time.Sleep(REJOIN_BACKOFF)
		}
	}
}

// dispatchEvents dispatch received events to dns & proxy goroutines
func (agent *Agent) dispatchEvents() {
	// send proxy info to dns proxy listener
	agent.Resolver.RecordChangeChan <- &nameserver.RecordChangeEvent{
		Change: "add",
		Record: nameserver.Record{
			Type:    nameserver.A,
			Ip:      agent.Config.Janitor.AdvertiseIP,
			IsProxy: true,
		},
	}

	for event := range agent.SerfServer.EventCh {
		userEvent, ok := event.(serf.UserEvent)
		if !ok {
			continue
		}

		var taskInfoEvent types.TaskInfoEvent
		err := json.Unmarshal(userEvent.Payload, &taskInfoEvent)
		if err != nil {
			logrus.Errorf("unmarshal taskInfoEvent go error: %s", err.Error())
			continue
		}

		if taskInfoEvent.GatewayEnabled {
			agent.Janitor.EventChan <- janitorTargetChangeEventFromTaskInfoEvent(
				userEvent.Name, &taskInfoEvent)
		}

		// Resolver only recongnize these two events
		if userEvent.Name == eventbus.EventTypeTaskHealthy ||
			userEvent.Name == eventbus.EventTypeTaskUnhealthy {
			agent.Resolver.RecordChangeChan <- recordChangeEventFromTaskInfoEvent(
				userEvent.Name, &taskInfoEvent)
		}
	}
}

// todo
func (agent *Agent) detectManagerLeader() (leaderAddr string, err error) {
	for _, managerAddr := range agent.Config.JoinAddrs {
		nodeRegistrationPath := managerAddr + "/ping"
		_, err := httpclient.NewDefaultClient().GET(context.TODO(), nodeRegistrationPath, nil, nil)
		if err != nil {
			logrus.Infof("register to %s got error: %s", nodeRegistrationPath, err.Error())
			continue
		}

		return managerAddr, nil
	}

	return "", errors.New("try join all managers are failed")
}

func (agent *Agent) watchManagerEvents(leaderAddr string) error {
	eventsDoesMatter := []string{
		eventbus.EventTypeTaskUnhealthy,
		eventbus.EventTypeTaskHealthy,
		eventbus.EventTypeTaskWeightChange,
	}

	eventsPath := fmt.Sprintf("http://%s/events?catchUp=true", leaderAddr)
	resp, err := http.Get(eventsPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		// skip blank line
		if line == SSE_BLANK_LINE {
			continue
		}

		if strings.HasPrefix(line, SSE_EVENT_PREFIX) {
			eventType := strings.TrimSpace(line[len(SSE_EVENT_PREFIX):len(line)])
			if !utils.SliceContains(eventsDoesMatter, eventType) {
				continue
			}

			// read next line of stream
			line, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			// if line is not data section
			if !strings.HasPrefix(line, SSE_DATA_PREFIX) {
				continue
			}

			agent.SerfServer.Publish(eventType, []byte(line[len(SSE_DATA_PREFIX):len(line)]))
		}
	}
}

func recordChangeEventFromTaskInfoEvent(eventType string, taskInfoEvent *types.TaskInfoEvent) *nameserver.RecordChangeEvent {
	resolverEvent := &nameserver.RecordChangeEvent{}
	if eventType == eventbus.EventTypeTaskHealthy {
		resolverEvent.Change = "add"
	} else {
		resolverEvent.Change = "del"
	}

	resolverEvent.Ip = taskInfoEvent.IP

	if taskInfoEvent.Mode == "replicates" {
		resolverEvent.Type = nameserver.SRV ^ nameserver.A
		resolverEvent.Port = fmt.Sprintf("%d", taskInfoEvent.Port)
	} else {
		resolverEvent.Type = nameserver.A
	}
	resolverEvent.Cluster = taskInfoEvent.ClusterID
	resolverEvent.RunAs = taskInfoEvent.RunAs
	resolverEvent.AppName = taskInfoEvent.AppName
	resolverEvent.InsName = taskInfoEvent.InsName
	resolverEvent.SlotID = fmt.Sprintf("%d", taskInfoEvent.SlotIndex)

	return resolverEvent
}

func janitorTargetChangeEventFromTaskInfoEvent(eventType string,
	taskInfoEvent *types.TaskInfoEvent) *janitor.TargetChangeEvent {

	janitorEvent := &janitor.TargetChangeEvent{}
	switch eventType {
	case eventbus.EventTypeTaskHealthy:
		janitorEvent.Change = "add"
	case eventbus.EventTypeTaskUnhealthy:
		janitorEvent.Change = "del"
	case eventbus.EventTypeTaskWeightChange:
		janitorEvent.Change = "change"
	default:
		return nil
	}

	janitorEvent.TaskIP = taskInfoEvent.IP
	janitorEvent.TaskPort = taskInfoEvent.Port
	janitorEvent.AppID = taskInfoEvent.AppID
	janitorEvent.PortName = taskInfoEvent.PortName
	janitorEvent.TaskPort = taskInfoEvent.Port
	janitorEvent.Weight = taskInfoEvent.Weight
	janitorEvent.TaskID = strings.ToLower(taskInfoEvent.TaskID)
	janitorEvent.AppVersion = taskInfoEvent.AppVersion
	janitorEvent.VersionID = taskInfoEvent.VersionID
	return janitorEvent
}
