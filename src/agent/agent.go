package agent

import (
	"strconv"

	"github.com/Dataman-Cloud/swan-janitor/src/janitor"
	"github.com/Dataman-Cloud/swan-resolver/nameserver"
	"github.com/Dataman-Cloud/swan/src/apiserver"
	"github.com/Dataman-Cloud/swan/src/swancontext"

	jconfig "github.com/Dataman-Cloud/swan-janitor/src/config"
	"golang.org/x/net/context"
)

type Agent struct {
	resolver *nameserver.Resolver

	janitorServer *janitor.JanitorServer

	CancelFunc context.CancelFunc
}

func New() (*Agent, error) {
	agent := &Agent{}

	dnsConfig := &nameserver.Config{
		Domain:   swancontext.Instance().Config.DNS.Domain,
		Listener: swancontext.Instance().Config.DNS.IP,
		Port:     swancontext.Instance().Config.DNS.Port,

		Resolvers:       swancontext.Instance().Config.DNS.Resolvers,
		ExchangeTimeout: swancontext.Instance().Config.DNS.ExchangeTimeout,
		SOARname:        swancontext.Instance().Config.DNS.SOARname,
		SOAMname:        swancontext.Instance().Config.DNS.SOAMname,
		SOASerial:       swancontext.Instance().Config.DNS.SOASerial,
		SOARefresh:      swancontext.Instance().Config.DNS.SOARefresh,
		SOARetry:        swancontext.Instance().Config.DNS.SOARetry,
		SOAExpire:       swancontext.Instance().Config.DNS.SOAExpire,
		RecurseOn:       swancontext.Instance().Config.DNS.RecurseOn,
		TTL:             swancontext.Instance().Config.DNS.TTL,
	}

	agent.resolver = nameserver.NewResolver(dnsConfig)

	jConfig := jconfig.DefaultConfig()
	jConfig.Listener.Mode = swancontext.Instance().Config.Janitor.ListenerMode
	jConfig.Listener.IP = swancontext.Instance().Config.Janitor.IP
	jConfig.Listener.DefaultPort = strconv.Itoa(swancontext.Instance().Config.Janitor.Port)
	jConfig.HttpHandler.Domain = swancontext.Instance().Config.Janitor.Domain
	agent.janitorServer = janitor.NewJanitorServer(jConfig)

	agentApi := &AgentApi{agent}
	apiserver.Install(swancontext.Instance().ApiServer, agentApi)

	return agent, nil
}

func (agent *Agent) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		resolverCtx, _ := context.WithCancel(ctx)
		errChan <- agent.resolver.Start(resolverCtx)
	}()

	go agent.janitorServer.Init().Run()

	// send proxy info to dns proxy listener
	rgEvent := &nameserver.RecordGeneratorChangeEvent{}
	rgEvent.Change = "add"
	rgEvent.Type = "a"
	rgEvent.Ip = swancontext.Instance().Config.Janitor.AdvertiseIP
	rgEvent.DomainPrefix = ""
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

func (agent *Agent) Stop(cancel context.CancelFunc) {
	//TODO resolver and janitor need stop
	//agent.resolver.Stop()
	//agent.janitorServer.Stop()
	cancel()
	return
}
