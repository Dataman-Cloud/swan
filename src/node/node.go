package node

import (
	"errors"
	"time"

	"github.com/Dataman-Cloud/swan/src/agent"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/event"
	"github.com/Dataman-Cloud/swan/src/manager"
	"github.com/Dataman-Cloud/swan/src/swancontext"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/utils/httpclient"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/twinj/uuid"
	"golang.org/x/net/context"
)

type Node struct {
	ID                string
	agent             *agent.Agent     // hold reference to agent, take function when in agent mode
	manager           *manager.Manager // hold a instance of manager, make logic taking place
	ctx               context.Context
	joinRetryInterval time.Duration
}

func NewNode(config config.SwanConfig, db *bolt.DB) (*Node, error) {
	// init swanconfig instance
	_ = swancontext.NewSwanContext(config, event.New())

	if !swancontext.IsManager() && !swancontext.IsAgent() {
		return nil, errors.New("node must be started with at least one role in [manager,agent]")
	}

	node := &Node{
		ID: uuid.NewV4().String(),

		joinRetryInterval: time.Second * 5,
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

	return node, nil
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
			n.JoinAsAgent()
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
	agentInfo := types.Node{
		ID:            n.ID,
		AdvertiseAddr: swanConfig.AdvertiseAddr,
	}

	for _, managerAddr := range swanConfig.SwanClusterAddrs {
		registerAddr := "http://" + managerAddr + config.API_PREFIX + "/manager/agents"
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
