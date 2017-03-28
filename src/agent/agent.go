package agent

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan-janitor/src"
	"github.com/Dataman-Cloud/swan-resolver/nameserver"
	"github.com/Dataman-Cloud/swan/src/config"
	eventbus "github.com/Dataman-Cloud/swan/src/event"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/utils"
	"github.com/Dataman-Cloud/swan/src/utils/httpclient"
	"github.com/Sirupsen/logrus"

	"golang.org/x/net/context"
)

const REJOIN_BACKOFF = 3 * time.Second

type Agent struct {
	resolver  *nameserver.Resolver
	janitor   *janitor.JanitorServer
	apiServer *AgentApiServer

	CancelFunc context.CancelFunc
	NodeInfo   types.Node
	JoinAddrs  []string
	Config     config.AgentConfig
}

func New(nodeID string, agentConf config.AgentConfig) (*Agent, error) {
	agent := &Agent{
		JoinAddrs: agentConf.JoinAddrs,
		Config:    agentConf,
	}

	nodeInfo := types.Node{
		ID:            nodeID,
		AdvertiseAddr: agentConf.AdvertiseAddr,
		ListenAddr:    agentConf.ListenAddr,
		Role:          types.RoleAgent,
	}
	agent.NodeInfo = nodeInfo

	agent.apiServer = NewAgentApiServer(agentConf.ListenAddr, agent)

	dnsConfig := &nameserver.Config{
		Domain:     agentConf.DNS.Domain,
		ListenAddr: agentConf.DNS.ListenAddr,

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

	agent.resolver = nameserver.NewResolver(dnsConfig)

	jConfig := janitor.DefaultConfig()
	jConfig.ListenAddr = agentConf.Janitor.ListenAddr
	jConfig.HttpHandler.Domain = agentConf.Janitor.Domain
	agent.janitor = janitor.NewJanitorServer(jConfig)

	return agent, nil
}

func (agent *Agent) StartAndJoin(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		errChan <- agent.start(ctx)
	}()

	go func() {
		for {
			leaderAddr, err := agent.join(ctx)
			if err != nil {
				logrus.Errorf("join to manager got error: %s", err.Error())
				time.Sleep(REJOIN_BACKOFF)
				continue
			}

			err = agent.watchEvents(leaderAddr)
			if err != nil {
				logrus.Errorf("watchEvents got error: %s", err.Error())
				logrus.Info("rejoin to cluster now")
				time.Sleep(REJOIN_BACKOFF)
			}
		}
	}()

	return <-errChan
}

func (agent *Agent) watchEvents(leaderAddr string) error {
	client := &http.Client{}
	// catchUp=true means get latest events
	eventsPath := fmt.Sprintf("http://%s/events?catchUp=true", leaderAddr)
	req, err := http.NewRequest("GET", eventsPath, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
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
		if len(line) == 0 {
			continue
		}

		eventsDoesMatter := []string{
			eventbus.EventTypeTaskUnhealthy,
			eventbus.EventTypeTaskHealthy,
		}

		if strings.HasPrefix(line, "event:") {
			eventType := strings.TrimSpace(strings.Split(line, ":")[1])
			if !utils.SliceContains(eventsDoesMatter, eventType) {
				continue
			}

			// read next line of stream
			line, err := reader.ReadString('\n')
			if err != nil {
				return err
			}

			// if line is not data section
			if !strings.HasPrefix(line, "data:") {
				continue
			}

			var taskInfoEvent types.TaskInfoEvent
			err = json.Unmarshal([]byte(line[5:len(line)]), &taskInfoEvent)
			if err != nil {
				logrus.Errorf("unmarshal taskInfoEvent go error: %s", err.Error())
				continue
			}

			agent.resolver.RecordGeneratorChangeChan() <- recordGeneratorChangeEventFromTaskInfoEvent(eventType, &taskInfoEvent)
			agent.janitor.SwanEventChan() <- janitorTargetgChangeEventFromTaskInfoEvent(eventType, &taskInfoEvent)
		}
	}
}

func recordGeneratorChangeEventFromTaskInfoEvent(eventType string, taskInfoEvent *types.TaskInfoEvent) *nameserver.RecordGeneratorChangeEvent {
	resolverEvent := &nameserver.RecordGeneratorChangeEvent{}
	if eventType == eventbus.EventTypeTaskHealthy {
		resolverEvent.Change = "add"
	} else {
		resolverEvent.Change = "del"
	}

	resolverEvent.Ip = taskInfoEvent.IP

	if taskInfoEvent.Mode == "replicates" {
		resolverEvent.Type = "srv"
		resolverEvent.Port = fmt.Sprintf("%d", taskInfoEvent.Port)
	} else {
		resolverEvent.Type = "a"
	}
	resolverEvent.DomainPrefix = strings.ToLower(strings.Replace(taskInfoEvent.TaskID, "-", ".", -1))
	return resolverEvent
}

func janitorTargetgChangeEventFromTaskInfoEvent(eventType string,
	taskInfoEvent *types.TaskInfoEvent) *janitor.TargetChangeEvent {

	janitorEvent := &janitor.TargetChangeEvent{}
	if eventType == eventbus.EventTypeTaskHealthy {
		janitorEvent.Change = "add"
	} else {
		janitorEvent.Change = "del"
	}

	janitorEvent.TaskIP = taskInfoEvent.IP
	janitorEvent.TaskPort = taskInfoEvent.Port
	janitorEvent.AppID = taskInfoEvent.AppID
	janitorEvent.PortName = taskInfoEvent.PortName
	janitorEvent.TaskPort = taskInfoEvent.Port
	janitorEvent.TaskID = strings.ToLower(taskInfoEvent.TaskID)
	return janitorEvent
}

func (agent *Agent) start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		errChan <- agent.apiServer.Start()
	}()

	go func() {
		resolverCtx, _ := context.WithCancel(ctx)
		errChan <- agent.resolver.Start(resolverCtx)
	}()

	go func() {
		err := agent.janitor.Init()
		if err != nil {
			errChan <- err
		}

		errChan <- agent.janitor.Run()
	}()

	// refactor later
	time.AfterFunc(time.Second, func() {
		// send proxy info to dns proxy listener
		rgEvent := &nameserver.RecordGeneratorChangeEvent{}
		rgEvent.Change = "add"
		rgEvent.Type = "a"
		rgEvent.Ip = agent.Config.Janitor.AdvertiseIP
		rgEvent.DomainPrefix = ""
		rgEvent.IsProxy = true
		agent.resolver.RecordGeneratorChangeChan() <- rgEvent
	})

	for {
		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (agent *Agent) Stop() {
	agent.CancelFunc()
	return
}

func (agent *Agent) join(ctx context.Context) (leaderAddr string, err error) {
	if len(agent.JoinAddrs) == 0 {
		return "", errors.New("start swan failed. Error: joinAddrs must be no empty")
	}

	for _, managerAddr := range agent.JoinAddrs {
		nodeRegistrationAddr := managerAddr + config.API_PREFIX + "/nodes"
		_, err := httpclient.NewDefaultClient().POST(context.TODO(), nodeRegistrationAddr, nil, agent.NodeInfo, nil)
		if err != nil {
			logrus.Infof("register to %s got error: %s", nodeRegistrationAddr, err.Error())
			continue
		}

		return managerAddr, nil
	}

	return "", errors.New("try join all managers are failed")
}
