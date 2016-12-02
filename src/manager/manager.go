package manager

import (
	"fmt"

	log "github.com/Dataman-Cloud/swan/src/context_logger"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/framework"
	"github.com/Dataman-Cloud/swan/src/manager/ipam"
	"github.com/Dataman-Cloud/swan/src/manager/ns"
	"github.com/Dataman-Cloud/swan/src/manager/raft"
	"github.com/Dataman-Cloud/swan/src/manager/sched"
	"github.com/Dataman-Cloud/swan/src/manager/store"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"
	"github.com/Dataman-Cloud/swan/src/util"
	"github.com/boltdb/bolt"

	"github.com/Sirupsen/logrus"
	events "github.com/docker/go-events"
	"golang.org/x/net/context"
)

type Manager struct {
	store     *store.Store
	apiserver *apiserver.ApiServer
	//proxyserver

	ipamAdapter *ipam.IpamAdapter
	resolver    *ns.Resolver
	sched       *sched.Sched
	raftNode    *raft.Node
	CancelFunc  context.CancelFunc

	framework *framework.Framework

	swanContext *swancontext.SwanContext
	config      util.SwanConfig
}

func New(config util.SwanConfig, db *bolt.DB) (*Manager, error) {
	manager := &Manager{
		config: config,
	}

	raftNode, err := raft.NewNode(config.Raft, db)
	if err != nil {
		logrus.Errorf("init raft node failed. Error: %s", err.Error())
		return nil, err
	}
	manager.raftNode = raftNode

	store := store.NewManagerStore(db, raftNode)

	manager.swanContext = &swancontext.SwanContext{
		Config: config,
		Store:  store,
		ApiServer: apiserver.NewApiServer(manager.config.HttpListener.TCPAddr,
			manager.config.HttpListener.UnixAddr),
	}

	manager.swanContext.Config.IPAM.StorePath = fmt.Sprintf(manager.config.IPAM.StorePath+"ipam.db.%d", config.Raft.RaftId)
	manager.ipamAdapter, err = ipam.New(manager.swanContext)
	if err != nil {
		logrus.Errorf("init ipam adapter failed. Error: ", err.Error())
		return nil, err
	}

	manager.resolver = ns.New(manager.config.DNS)
	manager.sched = sched.New(manager.config.Scheduler, manager.swanContext)
	manager.framework, err = framework.New(manager.config)
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

	raftCtx, _ := context.WithCancel(ctx)
	if err := manager.raftNode.StartRaft(raftCtx); err != nil {
		return err
	}

	if err := manager.raftNode.WaitForLeader(ctx); err != nil {
		return err
	}

	go func() {
		resolverCtx, _ := context.WithCancel(ctx)
		errCh <- manager.resolver.Start(resolverCtx)
	}()

	go func() {
		if err := manager.ipamAdapter.Start(); err != nil {
			errCh <- err
			return
		}

		managerRoute := NewRouter(manager)
		manager.swanContext.ApiServer.AppendRouter(managerRoute)

		errCh <- manager.swanContext.ApiServer.ListenAndServe()
	}()
	//go func() {
	//manager.framework.Start()
	//}()

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
				if manager.config.WithEngine == "sched" {
					sechedCtx, cancel := context.WithCancel(ctx)
					cancelFunc = cancel
					if err := manager.sched.Start(sechedCtx); err != nil {
						log.G(ctx).Error("Scheduler started unsuccessful")
						return
					}
					log.G(ctx).Info("Scheduler has been started")
				} else {
				}

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
