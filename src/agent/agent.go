package agent

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	janitor "github.com/Dataman-Cloud/swan-janitor/src"
	"github.com/Dataman-Cloud/swan-resolver/nameserver"
	"github.com/Dataman-Cloud/swan/src/config"
	eventbus "github.com/Dataman-Cloud/swan/src/event"
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

type Agent struct {
	Resolver   *nameserver.Resolver
	Janitor    *janitor.JanitorServer
	HttpServer *HttpServer
	SerfServer *SerfServer

	Config config.AgentConfig
}

func New(agentConf config.AgentConfig) (*Agent, error) {
	agent := &Agent{
		Config: agentConf,
	}

	resolverConfig := &nameserver.Config{
		Domain:     agentConf.DNS.Domain,
		ListenAddr: agentConf.DNS.ListenAddr,
		LogLevel:   agentConf.LogLevel,

		Resolvers:       agentConf.DNS.Resolvers,
		ExchangeTimeout: agentConf.DNS.ExchangeTimeout,
		SOARname:        agentConf.DNS.SOARname,
		SOAMname:        agentConf.DNS.SOAMname,
		SOASerial:       agentConf.DNS.SOASerial,
		SOARefresh:      agentConf.DNS.SOARefresh,
		SOARetry:        agentConf.DNS.SOARetry,
		SOAExpire:       agentConf.DNS.SOAExpire,
		RecurseOn:       agentConf.DNS.RecurseOn,
		TTL:             agentConf.DNS.TTL,
	}
	agent.Resolver = nameserver.NewResolver(resolverConfig)

	jConfig := janitor.DefaultConfig()
	jConfig.ListenAddr = agentConf.Janitor.ListenAddr
	jConfig.Domain = agentConf.Janitor.Domain
	jConfig.LogLevel = agentConf.LogLevel
	agent.Janitor = janitor.NewJanitorServer(jConfig)

	agent.HttpServer = NewHttpServer(agentConf.ListenAddr, agent)
	agent.SerfServer = NewSerfServer(agentConf.GossipListenAddr, agentConf.GossipJoinAddr)

	return agent, nil
}

func (agent *Agent) StartAndJoin(ctx context.Context) error {
	errChan := make(chan error)

	agentStartedCh := make(chan bool)
	go func() {
		err := agent.start(ctx, agentStartedCh)
		errChan <- err
	}()

	go func() {
		<-agentStartedCh
		for {
		JOIN_AGAIN:
			leaderAddr, err := agent.detectManagerLeader(ctx)
			if err != nil {
				logrus.Errorf("detect manager leader got error: %s", err.Error())
				time.Sleep(REJOIN_BACKOFF)
				goto JOIN_AGAIN
			}

			err = agent.watchManagerEvents(leaderAddr)
			if err != nil {
				logrus.Errorf("watchManagerEvents got error: %s", err.Error())
				time.Sleep(REJOIN_BACKOFF)
				goto JOIN_AGAIN
			}
		}
	}()

	return <-errChan
}

func (agent *Agent) start(ctx context.Context, started chan bool) error {
	errChan := make(chan error)
	var wg sync.WaitGroup
	wg.Add(4)
	resolverStarted, janitorStarted, httpStarted, serfStarted := make(chan bool),
		make(chan bool), make(chan bool), make(chan bool)

	go func() {
		for {
			select {
			case <-resolverStarted:
				wg.Done()
			case <-janitorStarted:
				wg.Done()
			case <-httpStarted:
				wg.Done()
			case <-serfStarted:
				wg.Done()
			}
		}
	}()

	go func() {
		resolverCtx, _ := context.WithCancel(ctx)
		errChan <- agent.Resolver.Start(resolverCtx, resolverStarted)
	}()

	go func() {
		janitorCtx, _ := context.WithCancel(ctx)
		errChan <- agent.Janitor.Start(janitorCtx, janitorStarted)
	}()

	go func() {
		serfCtx, _ := context.WithCancel(ctx)
		errChan <- agent.SerfServer.Start(serfCtx, serfStarted)
	}()

	go func() {
		httpServerCtx, _ := context.WithCancel(ctx)
		errChan <- agent.HttpServer.Start(httpServerCtx, httpStarted)
	}()

	go func() {
		wg.Wait()
		started <- true
		// send proxy info to dns proxy listener
		rgEvent := &nameserver.RecordChangeEvent{}
		rgEvent.Change = "add"
		rgEvent.Type = nameserver.A
		rgEvent.Ip = agent.Config.Janitor.AdvertiseIP

		rgEvent.IsProxy = true
		agent.Resolver.RecordChangeChan <- rgEvent

		for event := range agent.SerfServer.EventCh {
			userEvent, ok := event.(serf.UserEvent)
			if ok {
				var taskInfoEvent types.TaskInfoEvent
				err := json.Unmarshal(userEvent.Payload, &taskInfoEvent)
				if err != nil {
					logrus.Errorf("unmarshal taskInfoEvent go error: %s", err.Error())
				} else {
					logrus.Debugf("%+v", janitorTargetgChangeEventFromTaskInfoEvent(userEvent.Name, &taskInfoEvent))

					agent.Janitor.EventChan <- janitorTargetgChangeEventFromTaskInfoEvent(userEvent.Name, &taskInfoEvent)
					if userEvent.Name == eventbus.EventTypeTaskHealthy || userEvent.Name == eventbus.EventTypeTaskUnhealthy { // Resolver only recongnize these two events
						agent.Resolver.RecordChangeChan <- recordGeneratorChangeEventFromTaskInfoEvent(userEvent.Name, &taskInfoEvent)
					}
				}
			}
		}

	}()

	for {
		select {
		case err := <-errChan:
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// todo
func (agent *Agent) detectManagerLeader(ctx context.Context) (leaderAddr string, err error) {
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

func recordGeneratorChangeEventFromTaskInfoEvent(eventType string, taskInfoEvent *types.TaskInfoEvent) *nameserver.RecordChangeEvent {
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
	resolverEvent.SlotID = fmt.Sprintf("%d", taskInfoEvent.SlotIndex)

	return resolverEvent
}

func janitorTargetgChangeEventFromTaskInfoEvent(eventType string,
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
