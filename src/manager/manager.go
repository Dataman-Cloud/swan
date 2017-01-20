package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"

	"github.com/Dataman-Cloud/swan/src/config"
	log "github.com/Dataman-Cloud/swan/src/context_logger"
	"github.com/Dataman-Cloud/swan/src/event"
	swanevent "github.com/Dataman-Cloud/swan/src/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework"
	fstore "github.com/Dataman-Cloud/swan/src/manager/framework/store"
	"github.com/Dataman-Cloud/swan/src/manager/raft"
	rafttypes "github.com/Dataman-Cloud/swan/src/manager/raft/types"
	"github.com/Dataman-Cloud/swan/src/swancontext"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Dataman-Cloud/swan-janitor/src/upstream"
	"github.com/Dataman-Cloud/swan-resolver/nameserver"
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

	agents      map[string]types.Node
	managers    map[string]types.Node
	agentLock   sync.RWMutex
	managerLock sync.RWMutex

	janitorSubscriber  *event.JanitorSubscriber
	resolverSubscriber *event.DNSSubscriber
}

func New(db *bolt.DB) (*Manager, error) {
	manager := &Manager{
		criticalErrorChan: make(chan error, 1),
	}

	raftID, err := loadOrCreateRaftID(db)
	if err != nil {
		return nil, err
	}

	swanConfig := swancontext.Instance().Config
	raftNodeOpts := raft.NodeOptions{
		SwanNodeID:    swanConfig.NodeID,
		DataDir:       swanConfig.DataDir + "/" + swanConfig.NodeID,
		RaftID:        raftID,
		ListenAddr:    swanConfig.RaftListenAddr,
		AdvertiseAddr: swanConfig.AdvertiseAddr,
	}
	raftNode, err := raft.NewNode(raftNodeOpts, db)
	if err != nil {
		logrus.Errorf("init raft node failed. Error: %s", err.Error())
		return nil, err
	}
	manager.raftNode = raftNode

	frameworkStore := fstore.NewStore(db, raftNode)
	manager.framework, err = framework.New(frameworkStore, swancontext.Instance().ApiServer)
	if err != nil {
		logrus.Errorf("init framework failed. Error: ", err.Error())
		return nil, err
	}

	manager.resolverSubscriber = event.NewDNSSubscriber()
	manager.janitorSubscriber = event.NewJanitorSubscriber()

	return manager, nil
}

func loadOrCreateRaftID(db *bolt.DB) (uint64, error) {
	var raftID uint64
	tx, err := db.Begin(true)
	if err != nil {
		return raftID, err
	}
	defer tx.Commit()

	var (
		raftIDBukctName = []byte("raftnode")
		raftIDDataKey   = []byte("raftid")
	)
	raftIDBkt := tx.Bucket(raftIDBukctName)
	if raftIDBkt == nil {
		raftIDBkt, err = tx.CreateBucketIfNotExists(raftIDBukctName)
		if err != nil {
			return raftID, err
		}

		raftID = uint64(rand.Int63()) + 1
		if err := raftIDBkt.Put(raftIDDataKey, []byte(strconv.FormatUint(raftID, 10))); err != nil {
			return raftID, err
		}
		logrus.Infof("raft ID was not found create a new raftID %d", raftID)
		return raftID, nil
	} else {
		raftID_ := raftIDBkt.Get(raftIDDataKey)
		raftID, err = strconv.ParseUint(string(raftID_), 10, 64)
		if err != nil {
			return raftID, err
		}

		return raftID, nil
	}
}

func (manager *Manager) Stop(cancel context.CancelFunc) {
	cancel()
	return
}

func (manager *Manager) Start(ctx context.Context) error {
	if err := manager.LoadNodeData(); err != nil {
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

			ctx = log.WithLogger(ctx, logrus.WithField("raft_id", fmt.Sprintf("%x", manager.raftNode.Config.ID)))
			if newState == raft.IsLeader {
				log.G(ctx).Info("Now i become a leader !!!")

				eventBusCtx, _ := context.WithCancel(ctx)
				go func() {
					eventBusStarted = true
					log.G(eventBusCtx).Info("starting eventBus in leader.")
					manager.resolverSubscriber.Subscribe(swancontext.Instance().EventBus)
					manager.janitorSubscriber.Subscribe(swancontext.Instance().EventBus)
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
					manager.resolverSubscriber.Unsubscribe(swancontext.Instance().EventBus)
					manager.janitorSubscriber.Unsubscribe(swancontext.Instance().EventBus)
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
		case <-leaderChangeCh:
			//case leaderChangeEvent := <-leaderChangeCh:
			//var leaderAddr string
			//leader := leaderChangeEvent.(uint64)

			//// If leader was losted, this value is 0
			//if int(leader) == 0 {
			//	leaderAddr = ""
			//} else {
			//	leaderAddr = manager.clusterAddrs[int(leader)-1]
			//}

			//swancontext.Instance().ApiServer.UpdateLeaderAddr(leaderAddr)
			log.G(ctx).Info("Now leader is change to ", manager.raftNode.Config.ID)

		case <-ctx.Done():
			return
		}
	}
}

func (manager *Manager) LoadNodeData() error {
	nodes, err := manager.raftNode.GetNodes()
	if err != nil {
		return err
	}

	manager.agents = make(map[string]types.Node)
	for _, nodeMetadata := range nodes {
		node := types.Node{
			ID:            nodeMetadata.ID,
			AdvertiseAddr: nodeMetadata.AdvertiseAddr,
			ListenAddr:    nodeMetadata.ListenAddr,
			Role:          types.NodeRole(nodeMetadata.Role),
			Status:        nodeMetadata.Status,
			Labels:        nodeMetadata.Labels,
		}

		if node.IsAgent() {
			manager.AddAgentAcceptor(node)

			manager.agentLock.Lock()
			manager.agents[node.ID] = node
			manager.agentLock.Unlock()
		}
	}

	return nil
}

func (manager *Manager) AddAgent(agent types.Node) error {
	agentMetadata := &rafttypes.Node{
		ID:            agent.ID,
		AdvertiseAddr: agent.AdvertiseAddr,
		ListenAddr:    agent.ListenAddr,
		Status:        agent.Status,
		Labels:        agent.Labels,
		Role:          string(agent.Role),
	}

	storeActions := []*rafttypes.StoreAction{&rafttypes.StoreAction{
		Action: rafttypes.StoreActionKindCreate,
		Target: &rafttypes.StoreAction_Node{agentMetadata},
	}}

	if err := manager.raftNode.ProposeValue(context.TODO(), storeActions, nil); err != nil {
		return err
	}

	manager.AddAgentAcceptor(agent)

	manager.agentLock.Lock()
	manager.agents[agent.ID] = agent
	manager.agentLock.Unlock()

	go manager.SendAgentInitData(agent)

	return nil
}

func (manager *Manager) GetNodes() []types.Node {
	var nodes []types.Node
	manager.agentLock.RLock()
	for _, agent := range manager.agents {
		nodes = append(nodes, agent)
	}
	manager.agentLock.RUnlock()

	manager.managerLock.RLock()
	for _, m := range manager.managers {
		nodes = append(nodes, m)
	}
	manager.managerLock.RUnlock()

	return nodes
}

func (manager *Manager) GetNode(nodeID string) (types.Node, error) {
	manager.agentLock.RLock()
	node, ok := manager.agents[nodeID]
	manager.agentLock.RUnlock()
	if ok {
		return node, nil
	}

	manager.managerLock.RLock()
	node, ok = manager.managers[nodeID]
	manager.managerLock.RUnlock()
	if ok {
		return node, nil
	}

	return types.Node{}, errors.New("node not found")
}

func (manager *Manager) AddAgentAcceptor(agent types.Node) {
	resolverAcceptor := types.ResolverAcceptor{
		ID:         agent.ID,
		RemoteAddr: "http://" + agent.AdvertiseAddr + config.API_PREFIX + "/agent/resolver/event",
		Status:     agent.Status,
	}
	manager.resolverSubscriber.AddAcceptor(resolverAcceptor)

	janitorAcceptor := types.JanitorAcceptor{
		ID:         agent.ID,
		RemoteAddr: "http://" + agent.AdvertiseAddr + config.API_PREFIX + "/agent/janitor/event",
		Status:     agent.Status,
	}
	manager.janitorSubscriber.AddAcceptor(janitorAcceptor)
}

func (manager *Manager) SendAgentInitData(agent types.Node) {
	var resolverEvents []*nameserver.RecordGeneratorChangeEvent
	var janitorEvents []*upstream.TargetChangeEvent

	taskEvents := manager.framework.Scheduler.HealthyTaskEvents()

	for _, taskEvent := range taskEvents {
		resolverEvent, err := swanevent.BuildResolverEvent(taskEvent)
		if err == nil {
			resolverEvents = append(resolverEvents, resolverEvent)
		} else {
			logrus.Errorf("Build resolver event got error: %s", err.Error())
		}

		janitorEvent, err := swanevent.BuildJanitorEvent(taskEvent)
		if err == nil {
			janitorEvents = append(janitorEvents, janitorEvent)
		} else {
			logrus.Errorf("Build janitor event got error: %s", err.Error())
		}
	}

	resolverData, err := json.Marshal(resolverEvents)
	if err == nil {
		if err := swanevent.SendEventByHttp("http://"+agent.AdvertiseAddr+config.API_PREFIX+"/agent/resolver/init", "POST", resolverData); err != nil {
			logrus.Errorf("send resolver init data got error: %s", err.Error())
		}

	} else {
		logrus.Errorf("marshal resolver init data got error: %s", err.Error())
	}

	janitorData, err := json.Marshal(janitorEvents)
	if err == nil {
		if err := swanevent.SendEventByHttp("http://"+agent.AdvertiseAddr+config.API_PREFIX+"/agent/janitor/init", "POST", janitorData); err != nil {
			logrus.Errorf("send janitor init data got error: %s", err.Error())
		}
	} else {
		logrus.Errorf("marshal janitor init data got error: %s", err.Error())
	}
}
