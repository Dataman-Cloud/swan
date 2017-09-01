package zk

import (
	"fmt"
	"path"

	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
)

func (zk *ZKStore) CreateNode(vclusterId string, node *types.Node) error {
	bs, err := encode(node)
	if err != nil {
		return err
	}

	p := path.Join(keyVCluster, vclusterId, "nodes", node.IP)

	return zk.createAll(p, bs)
}

func (zk *ZKStore) GetNode(vId, nodeId string) (*types.Node, error) {
	p := path.Join(keyVCluster, vId, "nodes", nodeId)

	data, _, err := zk.get(p)
	if err != nil {
		if err == errNotExists {
			return nil, fmt.Errorf("node %s not exist", nodeId)
		}

		return nil, err
	}

	node := new(types.Node)
	if err := decode(data, node); err != nil {
		return nil, err
	}

	return node, nil
}

func (zk *ZKStore) UpdateNode(vId string, node *types.Node) error {
	bs, err := encode(node)
	if err != nil {
		return err
	}

	p := path.Join(keyVCluster, vId, "nodes", node.IP)

	return zk.set(p, bs)
}

func (zk *ZKStore) ListNodes(vId string) ([]*types.Node, error) {
	p := path.Join(keyVCluster, vId, "nodes")

	children, err := zk.list(p)
	if err != nil {
		return nil, err
	}

	nodes := make([]*types.Node, 0)
	for _, child := range children {
		node, err := zk.GetNode(vId, child)
		if err != nil {
			log.Errorf("ListNodes.GetNode error: %v", err)
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
