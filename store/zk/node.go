package zk

import (
	"fmt"
	"path"

	"github.com/Dataman-Cloud/swan/types"
)

func (zk *ZKStore) CreateNode(vclusterId string, node *types.Node) error {
	bs, err := encode(node)
	if err != nil {
		return err
	}

	p := path.Join(keyVCluster, vclusterId, "nodes", node.ID)

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

	p := path.Join(keyVCluster, vId, "nodes", node.ID)

	return zk.set(p, bs)
}
