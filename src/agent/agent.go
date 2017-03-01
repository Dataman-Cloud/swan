package agent

import (
	"errors"
	"strconv"
	"time"

	"github.com/Dataman-Cloud/swan-janitor/src"
	"github.com/Dataman-Cloud/swan-resolver/nameserver"
	"github.com/Dataman-Cloud/swan/src/apiserver"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/utils/httpclient"
	"github.com/Sirupsen/logrus"

	"golang.org/x/net/context"
)

const JoinRetryInterval = 5

type Agent struct {
	resolver *nameserver.Resolver

	janitorServer *janitor.JanitorServer

	apiServer *apiserver.ApiServer

	CancelFunc context.CancelFunc

	NodeInfo types.Node

	JoinAddrs []string

	Config config.AgentConfig
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

	agent.apiServer = apiserver.NewApiServer(agentConf.ListenAddr)

	dnsConfig := &nameserver.Config{
		Domain:   agentConf.DNS.Domain,
		Listener: agentConf.DNS.IP,
		Port:     agentConf.DNS.Port,

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
	jConfig.Listener.IP = agentConf.Janitor.IP
	jConfig.Listener.DefaultPort = strconv.Itoa(agentConf.Janitor.Port)
	jConfig.HttpHandler.Domain = agentConf.Janitor.Domain
	agent.janitorServer = janitor.NewJanitorServer(jConfig)

	agentApi := &AgentApi{agent}
	apiserver.Install(agent.apiServer, agentApi)

	return agent, nil
}

func (agent *Agent) JoinAndStart(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		errChan <- agent.Start(ctx)
	}()

	go func() {
		if err := agent.JoinToCluster(ctx); err != nil {
			errChan <- err
		}
	}()

	return <-errChan
}

func (agent *Agent) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		errChan <- agent.apiServer.Start()
	}()

	go func() {
		resolverCtx, _ := context.WithCancel(ctx)
		errChan <- agent.resolver.Start(resolverCtx)
	}()

	go agent.janitorServer.ServerInit().Run()

	// send proxy info to dns proxy listener
	rgEvent := &nameserver.RecordGeneratorChangeEvent{}
	rgEvent.Change = "add"
	rgEvent.Type = "a"
	//TODO(): better way is use janitor config replace agent config
	rgEvent.Ip = agent.Config.Janitor.AdvertiseIP
	rgEvent.DomainPrefix = ""
	rgEvent.IsProxy = true
	agent.resolver.RecordGeneratorChangeChan() <- rgEvent

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
	//TODO resolver and janitor need stop
	//agent.resolver.Stop()
	//agent.janitorServer.Stop()

	agent.CancelFunc()
	return
}

func (agent *Agent) JoinToCluster(ctx context.Context) error {
	tryJoinTimes := 1
	err := agent.join()
	if err != nil {
		logrus.Infof("join to swan cluster failed at %d times, retry after %d seconds", tryJoinTimes, JoinRetryInterval)
	} else {
		return nil
	}

	retryTicker := time.NewTicker(JoinRetryInterval * time.Second)
	defer retryTicker.Stop()

	for {
		select {
		case <-retryTicker.C:
			if tryJoinTimes >= 100 {
				return errors.New("join to swan cluster has been failed 100 times exit")
			}

			tryJoinTimes++

			err = agent.join()
			if err != nil {
				logrus.Infof("join to swan cluster failed at %d times, retry after %d seconds", tryJoinTimes, JoinRetryInterval)
			} else {
				return nil
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (agent *Agent) join() error {
	if len(agent.JoinAddrs) == 0 {
		return errors.New("start swan failed. Error: joinAddrs must be no empty")
	}

	for _, managerAddr := range agent.JoinAddrs {
		registerAddr := "http://" + managerAddr + config.API_PREFIX + "/nodes"
		_, err := httpclient.NewDefaultClient().POST(context.TODO(), registerAddr, nil, agent.NodeInfo, nil)
		if err != nil {
			logrus.Infof("register to %s got error: %s", registerAddr, err.Error())
			continue
		}

		logrus.Infof("join to swan cluster success with manager adderss %s", managerAddr)

		return nil
	}

	return errors.New("try join all managers are failed")
}
