package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

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

	"github.com/Dataman-Cloud/swan-janitor/src"
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

	criticalErrorChan chan error

	raftID uint64

	janitorSubscriber  *event.JanitorSubscriber
	resolverSubscriber *event.DNSSubscriber
}

func New(db *bolt.DB) (*Manager, error) {
	manager := &Manager{
		criticalErrorChan: make(chan error, 1),
	}

	swanConfig := swancontext.Instance().Config
	raftNodeOpts := raft.NodeOptions{
		SwanNodeID:    swanConfig.NodeID,
		DataDir:       swanConfig.DataDir + "/" + swanConfig.NodeID,
		ListenAddr:    swanConfig.RaftListenAddr,
		AdvertiseAddr: swanConfig.RaftAdvertiseAddr,
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

func (manager *Manager) Stop() {
	manager.CancelFunc()

	return
}

func (manager *Manager) Start(ctx context.Context, raftID uint64, raftPeers []types.Node, isNewCluster bool) error {
	manager.raftID = raftID

	if err := manager.LoadNodeData(); err != nil {
		return err
	}

	// when follower => leader or leader => follower
	leadershipCh, cancel := manager.raftNode.SubscribeLeadership()
	defer cancel()

	leadershipChangeEventCtx, _ := context.WithCancel(ctx)
	go manager.handleLeadershipEvents(leadershipChangeEventCtx, leadershipCh)

	leaderChangeCh, cancel := manager.raftNode.SubcribeLeaderChange()
	defer cancel()

	// when new leader was elected within cluster
	leaderChangeEventCtx, _ := context.WithCancel(ctx)
	go manager.handleLeaderChangeEvents(leaderChangeEventCtx, leaderChangeCh)

	raftCtx, _ := context.WithCancel(ctx)
	if err := manager.raftNode.StartRaft(raftCtx, manager.raftID, raftPeers, isNewCluster); err != nil {
		return err
	}

	// NOTICE: although WaitForLeader is returned, if call propose value as soon
	// there maybe return error: node losts leader status.
	// we should do propseValue in the handleLeadershipEvents go become leader event
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
	var once sync.Once
	for {
		select {
		case leadershipEvent := <-leadershipCh:
			// TODO lock it and if manager stop return
			newState := leadershipEvent.(raft.LeadershipState)

			ctx = log.WithLogger(ctx, logrus.WithField("raft_id", fmt.Sprintf("%x", manager.raftNode.Config.ID)))
			if newState == raft.IsLeader {
				log.G(ctx).Info("Now i become a leader !!!")

				once.Do(func() {
					// if manager is thie first node in the cluster, add itself to the managers map
					if swancontext.IsNewCluster() {
						swanConfig := swancontext.Instance().Config
						managerInfo := types.Node{
							ID:                swanConfig.NodeID,
							AdvertiseAddr:     swanConfig.AdvertiseAddr,
							ListenAddr:        swanConfig.ListenAddr,
							RaftListenAddr:    swanConfig.RaftListenAddr,
							RaftAdvertiseAddr: swanConfig.RaftAdvertiseAddr,
							Role:              types.NodeRole(swanConfig.Mode),
							RaftID:            manager.raftID,
						}

						if err := manager.AddManager(managerInfo); err != nil {
							manager.criticalErrorChan <- err
						}
					}
				})

				eventBusCtx, _ := context.WithCancel(ctx)
				go func() {
					log.G(eventBusCtx).Info("starting eventBus in leader.")

					eventBusStarted = true
					manager.resolverSubscriber.Subscribe(swancontext.Instance().EventBus)
					manager.janitorSubscriber.Subscribe(swancontext.Instance().EventBus)
					swancontext.Instance().EventBus.Start(ctx)
				}()

				frameworkCtx, _ := context.WithCancel(ctx)
				go func() {
					log.G(frameworkCtx).Info("starting framework in leader.")

					frameworkStarted = true
					manager.criticalErrorChan <- manager.framework.Start(frameworkCtx)
				}()

			} else if newState == raft.IsFollower {
				log.G(ctx).Info("Now i become a follower !!!")

				if eventBusStarted {
					manager.resolverSubscriber.Unsubscribe(swancontext.Instance().EventBus)
					manager.janitorSubscriber.Unsubscribe(swancontext.Instance().EventBus)
					swancontext.Instance().EventBus.Stop()
					eventBusStarted = false

					log.G(ctx).Info("eventBus has been stopped")

				}

				if frameworkStarted {
					log.G(ctx).Info("framework has been stopped")

					manager.framework.Stop()
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
			leaderRaftID := leaderChangeEvent.(uint64)

			manager.updateLeaderAddr(leaderRaftID)
		case <-ctx.Done():
			return
		}
	}
}

func (manager *Manager) updateLeaderAddr(leaderRaftID uint64) {
	var leaderAddr string

	// beacuse of the leader change event maybe published before all data has been sync.
	// maybe we can't find the leader node at first time, so need this loop for retry.
	// maintain a cluster membership maybe a better way like swarmkit:
	// https://github.com/docker/swarmkit/blob/master/manager/state/raft/membership/cluster.go
	for {
		// If leader was losted, this value is 0
		if int(leaderRaftID) == 0 {
			leaderAddr = ""
		} else {
			leaderNode, err := manager.findLeaderByRaftID(leaderRaftID)
			if err != nil {
				log.L.Warnf("update leaderAddr failed. Error: %s", err.Error())
				leaderAddr = ""
			} else {
				leaderAddr = leaderNode.AdvertiseAddr
			}

			swancontext.Instance().ApiServer.UpdateLeaderAddr(leaderAddr)
			log.L.Infof("Now leader is change to %x, leader advertise-addr: %s", leaderRaftID, leaderAddr)
		}

		// if leader was losted or the leader addresss was found return
		if int(leaderRaftID) == 0 || leaderAddr != "" {
			return
		}

		time.Sleep(time.Second)
	}
}

func (manager *Manager) findLeaderByRaftID(raftID uint64) (types.Node, error) {
	nodes := manager.GetNodes()
	for _, node := range nodes {
		if node.IsManager() && node.RaftID == raftID {
			return node, nil
		}
	}

	return types.Node{}, fmt.Errorf("can not find node which raftID is %x", raftID)
}

func (manager *Manager) LoadNodeData() error {
	nodes := manager.GetNodes()

	for _, node := range nodes {
		if node.IsAgent() {
			manager.AddAgentAcceptor(node)
		}
	}

	return nil
}

func (manager *Manager) AddAgent(agent types.Node) error {
	if err := manager.presistNodeData(agent); err != nil {
		return err
	}

	manager.AddAgentAcceptor(agent)

	go manager.SendAgentInitData(agent)

	return nil
}

func (manager *Manager) AddManager(m types.Node) error {
	return manager.presistNodeData(m)
}

func (manager *Manager) AddRaftNode(swanNode types.Node) error {
	if swanNode.RaftID == 0 {
		return errors.New("add raft node failed: raftID must not be 0")
	}

	swanNodes := manager.GetNodes()
	for _, n := range swanNodes {
		if n.RaftID == swanNode.RaftID && swanNode.ID != n.ID {
			return errors.New("add raft node failed: duplicate raftID")
		}
	}

	return manager.raftNode.AddMember(context.TODO(), swanNode)
}

func (manager *Manager) presistNodeData(node types.Node) error {
	nodeMetadata := &rafttypes.Node{
		ID:                node.ID,
		AdvertiseAddr:     node.AdvertiseAddr,
		ListenAddr:        node.ListenAddr,
		RaftListenAddr:    node.RaftListenAddr,
		RaftAdvertiseAddr: node.RaftAdvertiseAddr,
		Status:            node.Status,
		Labels:            node.Labels,
		RaftID:            node.RaftID,
		Role:              string(node.Role),
	}

	storeActions := []*rafttypes.StoreAction{&rafttypes.StoreAction{
		Action: rafttypes.StoreActionKindCreate,
		Target: &rafttypes.StoreAction_Node{nodeMetadata},
	}}

	return manager.raftNode.ProposeValue(context.TODO(), storeActions, nil)
}

func (manager *Manager) GetNodes() []types.Node {
	nodes := []types.Node{}

	nodes_, err := manager.raftNode.GetNodes()
	if err != nil {
		log.L.Errorf("get nodes failed. Error: %s", err.Error())
		return nodes
	}

	for _, nodeMetadata := range nodes_ {
		node := types.Node{
			ID:                nodeMetadata.ID,
			AdvertiseAddr:     nodeMetadata.AdvertiseAddr,
			ListenAddr:        nodeMetadata.ListenAddr,
			RaftListenAddr:    nodeMetadata.RaftListenAddr,
			RaftAdvertiseAddr: nodeMetadata.RaftAdvertiseAddr,
			Role:              types.NodeRole(nodeMetadata.Role),
			RaftID:            nodeMetadata.RaftID,
			Status:            nodeMetadata.Status,
			Labels:            nodeMetadata.Labels,
		}

		nodes = append(nodes, node)
	}

	return nodes
}

func (manager *Manager) GetNode(nodeID string) (types.Node, error) {
	nodes := manager.GetNodes()
	for _, node := range nodes {
		if node.ID == nodeID {
			return node, nil
		}
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
	var janitorEvents []*janitor.TargetChangeEvent

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
