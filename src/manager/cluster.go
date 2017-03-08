package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Dataman-Cloud/swan/src/config"
	log "github.com/Dataman-Cloud/swan/src/context_logger"
	swanevent "github.com/Dataman-Cloud/swan/src/event"
	rafttypes "github.com/Dataman-Cloud/swan/src/manager/raft/types"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/utils/httpclient"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

const JoinRetryInterval = 5

func (manager *Manager) AddNode(node types.Node) error {
	if err := manager.presistNodeData(node); err != nil {
		return err
	}

	if node.IsAgent() {
		manager.AddAgentAcceptor(node)

		go manager.SendAgentInitData(node)
	}

	// the first mixed node, node contains agent and leader, the leader node
	// can't add itself to raft-cluster, beacuse of it already in cluster
	if node.IsManager() && node.RaftID != manager.NodeInfo.RaftID {
		if err := manager.AddRaftNode(node); err != nil {
			return err
		}
	}

	return nil
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

func (manager *Manager) RemoveNode(node types.Node) error {
	if node.IsAgent() {
		manager.RemoveAgentAcceptor(node.ID)
	}

	if node.IsManager() {
		if err := manager.raftNode.RemoveMember(context.TODO(), node.RaftID); err != nil {
			return err
		}
	}

	storeActions := []*rafttypes.StoreAction{&rafttypes.StoreAction{
		Action: rafttypes.StoreActionKindRemove,
		Target: &rafttypes.StoreAction_Node{&rafttypes.Node{ID: node.ID}},
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
		nodes = append(nodes, converRaftTypeNodeToNode(*nodeMetadata))
	}

	return nodes
}

func (manager *Manager) GetManagers() []types.Node {
	nodes := manager.GetNodes()

	var managers []types.Node
	for _, node := range nodes {
		if node.IsManager() {
			managers = append(managers, node)
		}
	}

	return managers
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

func (manager *Manager) JoinToCluster(nodeInfo types.Node) ([]types.Node, error) {
	tryJoinTimes := 1
	existingNodes, err := manager.join(nodeInfo, manager.JoinAddrs)
	if err != nil {
		logrus.Infof("join to swan cluster failed at %d times, retry after %d seconds", tryJoinTimes, JoinRetryInterval)
	} else {
		return existingNodes, nil
	}

	retryTicker := time.NewTicker(JoinRetryInterval * time.Second)
	defer retryTicker.Stop()

	for {
		select {
		case <-retryTicker.C:
			if tryJoinTimes >= 100 {
				return nil, errors.New("join to swan cluster has been failed 100 times exit")
			}

			tryJoinTimes++

			existingNodes, err = manager.join(nodeInfo, manager.JoinAddrs)
			if err != nil {
				logrus.Infof("join to swan cluster failed at %d times, retry after %d seconds", tryJoinTimes, JoinRetryInterval)
			} else {
				return existingNodes, nil
			}
		}
	}
}

func (manager *Manager) join(nodeInfo types.Node, joinAddrs []string) ([]types.Node, error) {
	if len(joinAddrs) == 0 {
		return nil, errors.New("start swan failed. Error: joinAddrs must be no empty")
	}

	for _, managerAddr := range joinAddrs {
		registerAddr := managerAddr + config.API_PREFIX + "/nodes"
		resp, err := httpclient.NewDefaultClient().POST(context.TODO(), registerAddr, nil, nodeInfo, nil)
		if err != nil {
			logrus.Infof("register to %s got error: %s", registerAddr, err.Error())
			continue
		}

		var nodes []types.Node
		if err := json.Unmarshal(resp, &nodes); err != nil {
			logrus.Infof("register to %s got error: %s", registerAddr, err.Error())
			continue
		}

		var managerNodes []types.Node
		for _, existedNode := range nodes {
			if existedNode.IsManager() {
				managerNodes = append(managerNodes, existedNode)
			}
		}

		logrus.Infof("join to swan cluster success with manager adderss %s", managerAddr)

		return managerNodes, nil
	}

	return nil, errors.New("try join all managers are failed")
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

func (manager *Manager) presistNodeData(node types.Node) error {
	nodeMetadata := converNodeToRaftTypeNode(node)

	storeActions := []*rafttypes.StoreAction{&rafttypes.StoreAction{
		Action: rafttypes.StoreActionKindCreate,
		Target: &rafttypes.StoreAction_Node{&nodeMetadata},
	}}

	return manager.raftNode.ProposeValue(context.TODO(), storeActions, nil)
}

func (manager *Manager) AddAgentAcceptor(agent types.Node) {
	resolverAcceptor := types.ResolverAcceptor{
		ID:         agent.ID,
		RemoteAddr: agent.AdvertiseAddr + config.API_PREFIX + "/agent/resolver/event",
		Status:     agent.Status,
	}
	manager.resolverListener.AddAcceptor(resolverAcceptor)

	janitorAcceptor := types.JanitorAcceptor{
		ID:         agent.ID,
		RemoteAddr: agent.AdvertiseAddr + config.API_PREFIX + "/agent/janitor/event",
		Status:     agent.Status,
	}
	manager.janitorListener.AddAcceptor(janitorAcceptor)
}

func (manager *Manager) RemoveAgentAcceptor(agentID string) {
	manager.resolverListener.RemoveAcceptor(agentID)
	manager.janitorListener.RemoveAcceptor(agentID)
}

func (manager *Manager) SendAgentInitData(agent types.Node) {
	taskEvents := manager.framework.Scheduler.HealthyTaskEvents()

	if err := swanevent.SendEventByHttp(agent.AdvertiseAddr+config.API_PREFIX+"/agent/init", taskEvents); err != nil {
		logrus.Errorf("send resolver init data got error: %s", err.Error())
	}
}

func converRaftTypeNodeToNode(rafttypesNode rafttypes.Node) types.Node {
	return types.Node{
		ID:                rafttypesNode.ID,
		AdvertiseAddr:     rafttypesNode.AdvertiseAddr,
		ListenAddr:        rafttypesNode.ListenAddr,
		RaftListenAddr:    rafttypesNode.RaftListenAddr,
		RaftAdvertiseAddr: rafttypesNode.RaftAdvertiseAddr,
		Role:              types.NodeRole(rafttypesNode.Role),
		RaftID:            rafttypesNode.RaftID,
		Status:            rafttypesNode.Status,
		Labels:            rafttypesNode.Labels,
	}

}

func converNodeToRaftTypeNode(node types.Node) rafttypes.Node {
	return rafttypes.Node{
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
}
