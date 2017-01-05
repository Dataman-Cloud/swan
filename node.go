package main

import (
	"github.com/Dataman-Cloud/swan/src/agent"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager"
	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"

	"github.com/boltdb/bolt"
	"golang.org/x/net/context"
)

const (
	MODE_MANAGER = "manager"
	MODE_AGENT   = "agent"
	MODE_MIXED   = "mixed"
)

type Node struct {
	agent   *agent.Agent      // hold reference to agent, take function when in agent mode
	manager *manager.Manager  // hold a instance of manager, make logic taking place
	config  config.SwanConfig // swan config
	ctx     context.Context
}

func NewNode(config config.SwanConfig, db *bolt.DB) (*Node, error) {
	swanContext := swancontext.NewSwanContext(config, event.New())
	m, err := manager.New(*swanContext, db)
	if err != nil {
		return nil, err
	}

	a, err := agent.New(*swanContext)
	if err != nil {
		return nil, err
	}

	node := &Node{
		config:  config,
		manager: m,
		agent:   a,
	}

	return node, nil
}

func (n *Node) Start(ctx context.Context) error {
	errChan := make(chan error, 1)
	if n.config.Mode == MODE_MANAGER || n.config.Mode == MODE_MIXED {
		go func() {
			errChan <- n.runManager(ctx)
		}()
	}

	if n.config.Mode == MODE_AGENT || n.config.Mode == MODE_MIXED {
		go func() {
			errChan <- n.runAgent(ctx)
		}()
	}

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
