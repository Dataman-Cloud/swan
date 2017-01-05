package agent

import (
	"github.com/Dataman-Cloud/swan-janitor/src/janitor"
	"github.com/Dataman-Cloud/swan-resolver/nameserver"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"

	jconfig "github.com/Dataman-Cloud/swan-janitor/src/config"
	"golang.org/x/net/context"
)

type Agent struct {
	resolver           *nameserver.Resolver
	resolverSubscriber *event.DNSSubscriber

	janitorServer     *janitor.JanitorServer
	janitorSubscriber *event.JanitorSubscriber

	Config     config.SwanConfig
	CancelFunc context.CancelFunc
}

func New(swanContext swancontext.SwanContext) (*Agent, error) {
	agent := &Agent{
		Config: swanContext.Config,
	}

	if agent.Config.DNS.EnableDns {
		dnsConfig := &nameserver.Config{
			Domain:   agent.Config.DNS.Domain,
			Listener: agent.Config.DNS.Listener,
			Port:     agent.Config.DNS.Port,

			Resolvers:       agent.Config.DNS.Resolvers,
			ExchangeTimeout: agent.Config.DNS.ExchangeTimeout,
			SOARname:        agent.Config.DNS.SOARname,
			SOAMname:        agent.Config.DNS.SOAMname,
			SOASerial:       agent.Config.DNS.SOASerial,
			SOARefresh:      agent.Config.DNS.SOARefresh,
			SOARetry:        agent.Config.DNS.SOARetry,
			SOAExpire:       agent.Config.DNS.SOAExpire,
			RecurseOn:       agent.Config.DNS.RecurseOn,
			TTL:             agent.Config.DNS.TTL,
		}

		agent.resolver = nameserver.NewResolver(dnsConfig)
		agent.resolverSubscriber = event.NewDNSSubscriber(agent.resolver)
	}

	if agent.Config.Janitor.EnableProxy {
		jConfig := jconfig.DefaultConfig()
		jConfig.Listener.Mode = agent.Config.Janitor.ListenerMode
		jConfig.Listener.IP = agent.Config.Janitor.IP
		jConfig.Listener.DefaultPort = agent.Config.Janitor.Port
		jConfig.HttpHandler.Domain = agent.Config.Janitor.Domain
		agent.janitorServer = janitor.NewJanitorServer(jConfig)
		agent.janitorSubscriber = event.NewJanitorSubscriber(agent.janitorServer)
	}

	return agent, nil
}

func (agent *Agent) Start(ctx context.Context) error {
	errChan := make(chan error, 1)
	if agent.Config.DNS.EnableDns {
		go func() {
			resolverCtx, _ := context.WithCancel(ctx)
			agent.resolverSubscriber.Subscribe(swancontext.Instance().EventBus)
			errChan <- agent.resolver.Start(resolverCtx)
		}()
	}

	if agent.Config.Janitor.EnableProxy {
		agent.janitorSubscriber.Subscribe(swancontext.Instance().EventBus)
		go agent.janitorServer.Init().Run()
		// send proxy info to dns proxy listener
		if agent.Config.DNS.EnableDns {
			rgEvent := &nameserver.RecordGeneratorChangeEvent{}
			rgEvent.Change = "add"
			rgEvent.Type = "a"
			rgEvent.Ip = agent.Config.Janitor.IP
			rgEvent.DomainPrefix = ""
			agent.resolver.RecordGeneratorChangeChan() <- rgEvent
		}
	}

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
