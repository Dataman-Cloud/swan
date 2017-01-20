package node

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/Dataman-Cloud/swan/src/agent"
	"github.com/Dataman-Cloud/swan/src/apiserver"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/event"
	"github.com/Dataman-Cloud/swan/src/manager"
	"github.com/Dataman-Cloud/swan/src/swancontext"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/utils/httpclient"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/twinj/uuid"
	"golang.org/x/net/context"
)

const (
	NodeIDFileName = "/ID"
)

type Node struct {
	ID                string
	agent             *agent.Agent     // hold reference to agent, take function when in agent mode
	manager           *manager.Manager // hold a instance of manager, make logic taking place
	ctx               context.Context
	joinRetryInterval time.Duration
}

func NewNode(config config.SwanConfig) (*Node, error) {
	nodeID, err := loadOrCreateNodeID(config)
	if err != nil {
		return nil, err
	}

	// init swanconfig instance
	config.NodeID = nodeID
	_ = swancontext.NewSwanContext(config, event.New())

	if !swancontext.IsManager() && !swancontext.IsAgent() {
		return nil, errors.New("node must be started with at least one role in [manager,agent]")
	}

	node := &Node{
		ID:                nodeID,
		joinRetryInterval: time.Second * 5,
	}

	os.MkdirAll(config.DataDir+"/"+nodeID, 0700)

	db, err := bolt.Open(config.DataDir+"/"+nodeID+"/swan.db", 0600, nil)
	if err != nil {
		logrus.Errorf("Init store engine failed:%s", err)
		return nil, err
	}

	if swancontext.IsManager() {
		m, err := manager.New(db)
		if err != nil {
			return nil, err
		}
		node.manager = m
	}

	if swancontext.IsAgent() {
		a, err := agent.New()
		if err != nil {
			return nil, err
		}
		node.agent = a
	}

	nodeApi := &NodeApi{node}
	apiserver.Install(swancontext.Instance().ApiServer, nodeApi)

	return node, nil
}

func loadOrCreateNodeID(swanConfig config.SwanConfig) (string, error) {
	nodeIDFile := swanConfig.DataDir + NodeIDFileName
	if !fileutil.Exist(nodeIDFile) {
		os.MkdirAll(swanConfig.DataDir, 0700)

		nodeID := uuid.NewV4().String()
		idFile, err := os.OpenFile(nodeIDFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return "", err
		}

		if _, err = idFile.WriteString(nodeID); err != nil {
			return "", err
		}

		logrus.Infof("starting swan node, ID file was not found started with  new ID: %s", nodeID)
		return nodeID, nil

	} else {
		idFile, err := os.Open(nodeIDFile)
		if err != nil {
			return "", err
		}

		nodeID, err := ioutil.ReadAll(idFile)
		if err != nil {
			return "", err
		}

		logrus.Infof("starting swan node, ID file was found started with ID: %s", string(nodeID))
		return string(nodeID), nil
	}
}

func (n *Node) Start(ctx context.Context) error {
	errChan := make(chan error, 1)
	if swancontext.IsManager() {
		go func() {
			errChan <- n.runManager(ctx)
		}()
	}

	if swancontext.IsAgent() {
		go func() {
			errChan <- n.runAgent(ctx)
		}()

		go func() {
			err := n.JoinAsAgent()
			if err != nil {
				errChan <- err
			}
		}()
	}

	go func() {
		errChan <- swancontext.Instance().ApiServer.Start()
	}()

	for {
		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (n *Node) runAgent(ctx context.Context) error {
	agentCtx, cancel := context.WithCancel(ctx)
	n.agent.CancelFunc = cancel
	return n.agent.Start(agentCtx)
}

func (n *Node) runManager(ctx context.Context) error {
	managerCtx, cancel := context.WithCancel(ctx)
	n.manager.CancelFunc = cancel
	return n.manager.Start(managerCtx)
}

func (n *Node) stopManager() {
	n.agent.Stop(n.agent.CancelFunc)
	n.manager.Stop(n.manager.CancelFunc)
}

func (n *Node) JoinAsAgent() error {
	swanConfig := swancontext.Instance().Config
	if len(swanConfig.JoinAddrs) == 0 {
		return errors.New("start agent failed. Error: joinAddrs must be no empty")
	}

	agentInfo := types.Node{
		ID:            n.ID,
		AdvertiseAddr: swanConfig.AdvertiseAddr,
		ListenAddr:    swanConfig.ListenAddr,
		Role:          types.NodeRole(swanConfig.Mode),
	}

	for _, managerAddr := range swanConfig.JoinAddrs {
		registerAddr := "http://" + managerAddr + config.API_PREFIX + "/nodes"
		_, err := httpclient.NewDefaultClient().POST(context.TODO(), registerAddr, nil, agentInfo, nil)
		if err != nil {
			logrus.Errorf("register to %s got error: %s", registerAddr, err.Error())
		}

		if err == nil {
			logrus.Infof("agent register to manager success with managerAddr: %s", managerAddr)
			return nil
		}
	}

	time.Sleep(n.joinRetryInterval)
	n.JoinAsAgent()
	return nil
}
