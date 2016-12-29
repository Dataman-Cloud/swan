package manager

import (
	"fmt"

	"github.com/Dataman-Cloud/swan-resolver/nameserver"
	"github.com/Dataman-Cloud/swan/src/config"
	log "github.com/Dataman-Cloud/swan/src/context_logger"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework"
	fstore "github.com/Dataman-Cloud/swan/src/manager/framework/store"
	"github.com/Dataman-Cloud/swan/src/manager/raft"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"

	jconfig "github.com/Dataman-Cloud/swan-janitor/src/config"
	"github.com/Dataman-Cloud/swan-janitor/src/janitor"
	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	events "github.com/docker/go-events"
	"golang.org/x/net/context"
)

type Manager struct {
	resolver           *nameserver.Resolver
	resolverSubscriber *event.DNSSubscriber

	janitorServer     *janitor.JanitorServer
	janitorSubscriber *event.JanitorSubscriber

	raftNode   *raft.Node
	CancelFunc context.CancelFunc

	framework *framework.Framework

	config    config.SwanConfig
	cluster   []string
	apiserver *apiserver.ApiServer

	criticalErrorChan chan error
}

func New(config config.SwanConfig, db *bolt.DB) (*Manager, error) {
	manager := &Manager{
		config:            config,
		criticalErrorChan: make(chan error, 1),
	}

	manager.apiserver = apiserver.NewApiServer(config.HttpListener.TCPAddr, config.HttpListener.UnixAddr)

	raftNode, err := raft.NewNode(config.Raft, db)
	if err != nil {
		logrus.Errorf("init raft node failed. Error: %s", err.Error())
		return nil, err
	}
	manager.raftNode = raftNode

	swancontext.NewSwanContext(config, event.New())

	manager.cluster = manager.config.SwanCluster

	if manager.config.DNS.EnableDns {
		dnsConfig := &nameserver.Config{
			Domain:   manager.config.DNS.Domain,
			Listener: manager.config.DNS.Listener,
			Port:     manager.config.DNS.Port,

			Resolvers:       manager.config.DNS.Resolvers,
			ExchangeTimeout: manager.config.DNS.ExchangeTimeout,
			SOARname:        manager.config.DNS.SOARname,
			SOAMname:        manager.config.DNS.SOAMname,
			SOASerial:       manager.config.DNS.SOASerial,
			SOARefresh:      manager.config.DNS.SOARefresh,
			SOARetry:        manager.config.DNS.SOARetry,
			SOAExpire:       manager.config.DNS.SOAExpire,
			RecurseOn:       manager.config.DNS.RecurseOn,
			TTL:             manager.config.DNS.TTL,
		}

		manager.resolver = nameserver.NewResolver(dnsConfig)
		manager.resolverSubscriber = event.NewDNSSubscriber(manager.resolver)
	}

	if manager.config.Janitor.EnableProxy {
		jConfig := jconfig.DefaultConfig()
		jConfig.Listener.Mode = manager.config.Janitor.ListenerMode
		jConfig.Listener.IP = manager.config.Janitor.IP
		jConfig.Listener.DefaultPort = manager.config.Janitor.Port
		jConfig.HttpHandler.Domain = manager.config.Janitor.Domain
		manager.janitorServer = janitor.NewJanitorServer(jConfig)
		manager.janitorSubscriber = event.NewJanitorSubscriber(manager.janitorServer)
	}

	frameworkStore := fstore.NewStore(db, raftNode)
	manager.framework, err = framework.New(frameworkStore, manager.apiserver)
	if err != nil {
		logrus.Errorf("init framework failed. Error: ", err.Error())
		return nil, err
	}

	return manager, nil
}

func (manager *Manager) Stop(cancel context.CancelFunc) {
	cancel()
	return
}

func (manager *Manager) Start(ctx context.Context) error {
	leadershipCh, QueueCancel := manager.raftNode.SubscribeLeadership()
	defer QueueCancel()

	eventCtx, _ := context.WithCancel(ctx)
	go manager.handleLeadershipEvents(eventCtx, leadershipCh)

	leaderChangeCh, leaderChangeQueueCancel := manager.raftNode.SubcribeLeaderChange()
	defer leaderChangeQueueCancel()

	leaderChangeEventCtx, _ := context.WithCancel(ctx)
	go manager.handleLeaderChangeEvents(leaderChangeEventCtx, leaderChangeCh)

	raftCtx, _ := context.WithCancel(ctx)
	if err := manager.raftNode.StartRaft(raftCtx); err != nil {
		return err
	}

	if err := manager.raftNode.WaitForLeader(ctx); err != nil {
		return err
	}

	if manager.config.DNS.EnableDns {
		go func() {
			resolverCtx, _ := context.WithCancel(ctx)
			manager.resolverSubscriber.Subscribe(swancontext.Instance().EventBus)
			manager.criticalErrorChan <- manager.resolver.Start(resolverCtx)
		}()
	}

	if manager.config.Janitor.EnableProxy {
		manager.janitorSubscriber.Subscribe(swancontext.Instance().EventBus)
		go manager.janitorServer.Init().Run()
		// send proxy info to dns proxy listener
		if manager.config.DNS.EnableDns {
			rgEvent := &nameserver.RecordGeneratorChangeEvent{}
			rgEvent.Change = "add"
			rgEvent.Type = "a"
			rgEvent.Ip = manager.config.Janitor.IP
			rgEvent.DomainPrefix = ""
			manager.resolver.RecordGeneratorChangeChan() <- rgEvent
		}
	}

	go manager.apiserver.Start()

	for {
		select {
		case err := <-manager.criticalErrorChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (manager *Manager) handleLeadershipEvents(ctx context.Context, leadershipCh chan events.Event) {
	var eventBusStarted, frameworkStarted bool
	for {
		select {
		case leadershipEvent := <-leadershipCh:
			// TODO lock it and if manager stop return
			newState := leadershipEvent.(raft.LeadershipState)

			ctx = log.WithLogger(ctx, logrus.WithField("raft_id", fmt.Sprintf("%x", manager.config.Raft.RaftId)))
			if newState == raft.IsLeader {
				log.G(ctx).Info("Now i become a leader !!!")

				eventBusCtx, _ := context.WithCancel(ctx)
				go func() {
					eventBusStarted = true
					log.G(eventBusCtx).Info("starting eventBus in leader.")
					swancontext.Instance().EventBus.Start(ctx)
				}()

				frameworkCtx, _ := context.WithCancel(ctx)
				go func() {
					frameworkStarted = true
					log.G(frameworkCtx).Info("starting framework in leader.")
					manager.criticalErrorChan <- manager.framework.Start(frameworkCtx)
				}()

			} else if newState == raft.IsFollower {
				log.G(ctx).Info("Now i become a follower !!!")

				if eventBusStarted {
					swancontext.Instance().EventBus.Stop()
					log.G(ctx).Info("eventBus has been stopped")
					eventBusStarted = false
				}

				if frameworkStarted {
					manager.framework.Stop()
					log.G(ctx).Info("framework has been stopped")
					frameworkStarted = false
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (manager *Manager) handleLeaderChangeEvents(ctx context.Context, leaderChangeCh chan events.Event) {
	for {
		select {
		case leaderChangeEvent := <-leaderChangeCh:
			var leaderAddr string
			leader := leaderChangeEvent.(uint64)

			// If leader was losted, this value is 0
			if int(leader) == 0 {
				leaderAddr = ""
			} else {
				leaderAddr = manager.cluster[int(leader)-1]
			}

			manager.apiserver.UpdateLeaderAddr(leaderAddr)
			log.G(ctx).Info("Now leader is change to ", leaderAddr)

		case <-ctx.Done():
			return
		}
	}
}
