package manager

import (
	"fmt"

	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
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

	swanContext *swancontext.SwanContext
	config      util.SwanConfig
}

func New(config util.SwanConfig, db *bolt.DB) (*Manager, error) {
	manager := &Manager{
		config: config,
	}

	raftNode, err := raft.NewNode(config.Raft, db)
	if err != nil {
		logrus.Errorf("inti raft node failed. Error: %s", err.Error())
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

		errCh <- manager.swanContext.ApiServer.ListenAndServe()
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
			if newState == raft.IsLeader {
				sechedCtx, cancel := context.WithCancel(ctx)
				cancelFunc = cancel
				manager.sched.Start(sechedCtx)
			} else if newState == raft.IsFollower {
				if cancelFunc != nil {
					cancelFunc()
				}
				cancelFunc = nil
			}
		case <-ctx.Done():
			return
		}
	}
}
