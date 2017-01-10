package manager

import (
	"fmt"
	"sync"

	"github.com/Dataman-Cloud/swan/src/apiserver"
	"github.com/Dataman-Cloud/swan/src/config"
	log "github.com/Dataman-Cloud/swan/src/context_logger"
	"github.com/Dataman-Cloud/swan/src/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework"
	fstore "github.com/Dataman-Cloud/swan/src/manager/framework/store"
	"github.com/Dataman-Cloud/swan/src/manager/raft"
	rafttypes "github.com/Dataman-Cloud/swan/src/manager/raft/types"
	"github.com/Dataman-Cloud/swan/src/swancontext"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	events "github.com/docker/go-events"
	"golang.org/x/net/context"
)

type Manager struct {
	raftNode   *raft.Node
	CancelFunc context.CancelFunc

	framework *framework.Framework

	clusterAddrs []string

	criticalErrorChan chan error

	agents    map[string]types.Agent
	agentLock sync.RWMutex

	janitorSubscriber  *event.JanitorSubscriber
	resolverSubscriber *event.DNSSubscriber
}

func New(db *bolt.DB) (*Manager, error) {
	manager := &Manager{
		criticalErrorChan: make(chan error, 1),
	}

	raftNode, err := raft.NewNode(swancontext.Instance().Config.Raft, db)
	if err != nil {
		logrus.Errorf("init raft node failed. Error: %s", err.Error())
		return nil, err
	}
	manager.raftNode = raftNode

	manager.clusterAddrs = swancontext.Instance().Config.SwanClusterAddrs

	frameworkStore := fstore.NewStore(db, raftNode)
	manager.framework, err = framework.New(frameworkStore, swancontext.Instance().ApiServer)
	if err != nil {
		logrus.Errorf("init framework failed. Error: ", err.Error())
		return nil, err
	}

	managerApi := &ManagerApi{manager}
	apiserver.Install(swancontext.Instance().ApiServer, managerApi)

	manager.resolverSubscriber = event.NewDNSSubscriber()
	manager.janitorSubscriber = event.NewJanitorSubscriber()

	return manager, nil
}

func (manager *Manager) Stop(cancel context.CancelFunc) {
	cancel()
	return
}

func (manager *Manager) Start(ctx context.Context) error {
	if err := manager.LoadAgentData(); err != nil {
		return err
	}

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

			ctx = log.WithLogger(ctx, logrus.WithField("raft_id", fmt.Sprintf("%x", swancontext.Instance().Config.Raft.RaftId)))
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
				leaderAddr = manager.clusterAddrs[int(leader)-1]
			}

			swancontext.Instance().ApiServer.UpdateLeaderAddr(leaderAddr)
			log.G(ctx).Info("Now leader is change to ", leaderAddr)

		case <-ctx.Done():
			return
		}
	}
}

func (manager *Manager) LoadAgentData() error {
	agents, err := manager.raftNode.GetAgents()
	if err != nil {
		return err
	}

	manager.agents = make(map[string]types.Agent)
	for _, agentMetadata := range agents {
		agent := types.Agent{
			ID:         agentMetadata.ID,
			RemoteAddr: agentMetadata.RemoteAddr,
			Status:     agentMetadata.Status,
			Labels:     agentMetadata.Labels,
		}

		manager.AddAgentAcceptor(agent)

		manager.agentLock.Lock()
		manager.agents[agent.ID] = agent
		manager.agentLock.Unlock()
	}

	return nil
}

func (manager *Manager) AddAgent(agent types.Agent) error {
	agentMetadata := &rafttypes.Agent{
		ID:         agent.ID,
		RemoteAddr: agent.RemoteAddr,
		Status:     agent.Status,
		Labels:     agent.Labels,
	}

	storeActions := []*rafttypes.StoreAction{&rafttypes.StoreAction{
		Action: rafttypes.StoreActionKindCreate,
		Target: &rafttypes.StoreAction_Agent{agentMetadata},
	}}

	if err := manager.raftNode.ProposeValue(context.TODO(), storeActions, nil); err != nil {
		return err
	}

	manager.AddAgentAcceptor(agent)

	manager.agentLock.Lock()
	manager.agents[agent.ID] = agent
	manager.agentLock.Unlock()
	return nil
}

func (manager *Manager) GetAgents() map[string]types.Agent {
	return manager.agents
}

func (manager *Manager) UpdateAgent(agent types.Agent) error {
	return nil
}

func (manager *Manager) GetAgent(agentID string) types.Agent {
	manager.agentLock.RLock()
	defer manager.agentLock.RUnlock()
	return manager.agents[agentID]
}

func (manager *Manager) DeleteAgent(agentID string) error {
	agentMetadata := &rafttypes.Agent{
		ID: agentID,
	}

	storeActions := []*rafttypes.StoreAction{&rafttypes.StoreAction{
		Action: rafttypes.StoreActionKindRemove,
		Target: &rafttypes.StoreAction_Agent{agentMetadata},
	}}

	if err := manager.raftNode.ProposeValue(context.TODO(), storeActions, nil); err != nil {
		return err
	}

	manager.agentLock.Lock()
	delete(manager.agents, agentID)
	manager.agentLock.Unlock()
	return nil
}

func (manager *Manager) AddAgentAcceptor(agent types.Agent) {
	resolverAcceptor := types.ResolverAcceptor{
		ID:         agent.ID,
		RemoteAddr: "http://" + agent.RemoteAddr + config.API_PREFIX + "/agent/resolver/event",
		Status:     agent.Status,
	}
	manager.resolverSubscriber.AddAcceptor(resolverAcceptor)

	janitorAcceptor := types.JanitorAcceptor{
		ID:         agent.ID,
		RemoteAddr: "http://" + agent.RemoteAddr + config.API_PREFIX + "/agent/janitor/event",
		Status:     agent.Status,
	}
	manager.janitorSubscriber.AddAcceptor(janitorAcceptor)
}
