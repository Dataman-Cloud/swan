package etcd

import (
	"github.com/Dataman-Cloud/swan/types"
)

func (s *EtcdStore) CreateNode(vclusterId string, node *types.Node) error {
	return nil
}

func (s *EtcdStore) GetNode(vId, nodeId string) (*types.Node, error) {
	return nil, nil
}

func (s *EtcdStore) UpdateNode(vId string, node *types.Node) error {
	return nil
}

func (s *EtcdStore) ListNodes(vId string) ([]*types.Node, error) {
	return nil, nil
}

func (s *EtcdStore) DeleteNode(name, nodeIp string) error {
	return nil
}
