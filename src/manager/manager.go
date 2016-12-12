package manager

import (
	"fmt"

	"github.com/Dataman-Cloud/swan-resolver/nameserver"
	"github.com/Dataman-Cloud/swan/src/config"
	log "github.com/Dataman-Cloud/swan/src/context_logger"
	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework"
	fstore "github.com/Dataman-Cloud/swan/src/manager/framework/store"
	"github.com/Dataman-Cloud/swan/src/manager/ipam"
	"github.com/Dataman-Cloud/swan/src/manager/raft"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"
	"github.com/boltdb/bolt"

	"github.com/Sirupsen/logrus"
	events "github.com/docker/go-events"
	"golang.org/x/net/context"
)

type Manager struct {
	ipamAdapter *ipam.IpamAdapter

	resolver           *nameserver.Resolver
	resolverSubscriber *event.DNSSubscriber

	raftNode   *raft.Node
	CancelFunc context.CancelFunc
	eventBus   *event.EventBus

	framework *framework.Framework

	swanContext *swancontext.SwanContext
	config      config.SwanConfig
	cluster     []string
}

func New(config config.SwanConfig, db *bolt.DB) (*Manager, error) {
	manager := &Manager{
		config: config,
	}

	raftNode, err := raft.NewNode(config.Raft, db)
	if err != nil {
		logrus.Errorf("init raft node failed. Error: %s", err.Error())
		return nil, err
	}
	manager.raftNode = raftNode

	manager.eventBus = event.New()

	manager.swanContext = &swancontext.SwanContext{
		Config:   config,
		EventBus: manager.eventBus,
	}

	manager.swanContext.Config.IPAM.StorePath = fmt.Sprintf(manager.config.IPAM.StorePath+"ipam.db.%d", config.Raft.RaftId)
	manager.ipamAdapter, err = ipam.New(manager.swanContext)
	if err != nil {
		logrus.Errorf("init ipam adapter failed. Error: ", err.Error())
		return nil, err
	}

	manager.cluster = manager.config.SwanCluster
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

	frameworkStore := fstore.NewStore(db, raftNode)
	manager.framework, err = framework.New(manager.swanContext, manager.config, frameworkStore)
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
	errCh := make(chan error)
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

	go func() {
		resolverCtx, _ := context.WithCancel(ctx)
		manager.resolverSubscriber.Subscribe(manager.eventBus)
		errCh <- manager.resolver.Start(resolverCtx)
	}()

	go func() {
		if err := manager.ipamAdapter.Start(); err != nil {
			errCh <- err
			return
		}
	}()

	return <-errCh
}

func (manager *Manager) handleLeadershipEvents(ctx context.Context, leadershipCh chan events.Event) {
	for {
		select {
		case leadershipEvent := <-leadershipCh:
			// TODO lock it and if manager stop return
			newState := leadershipEvent.(raft.LeadershipState)

			var cancelFunc context.CancelFunc
			ctx = log.WithLogger(ctx, logrus.WithField("raft_id", fmt.Sprintf("%x", manager.config.Raft.RaftId)))
			if newState == raft.IsLeader {
				log.G(ctx).Info("Now i become a leader !!!")
				// TODO
				go func() {
					manager.eventBus.Start()
				}()

				frameworkCtx, _ := context.WithCancel(ctx)
				manager.framework.Start(frameworkCtx)

			} else if newState == raft.IsFollower {
				log.G(ctx).Info("Now i become a follower !!!")
				if cancelFunc != nil {
					cancelFunc()
					log.G(ctx).Info("Scheduler has been stopped")
				}
				cancelFunc = nil
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
			leader := leaderChangeEvent.(uint64)

			leaderAddr := manager.cluster[int(leader)-1]
			log.G(ctx).Info("Now leader is change to ", leaderAddr)

		case <-ctx.Done():
			return
		}
	}
}
