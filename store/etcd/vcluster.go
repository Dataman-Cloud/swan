package etcd

import (
	"github.com/Dataman-Cloud/swan/types"
)

func (etcd *EtcdStore) ListVClusters() ([]*types.VCluster, error) {
	return nil, nil
}

func (etcd *EtcdStore) CreateVCluster(*types.VCluster) error {
	return nil
}

func (etcd *EtcdStore) GetVCluster(vclusterId string) (*types.VCluster, error) {
	return nil, nil
}

func (etcd *EtcdStore) VClusterExists(name string) bool {
	return false
}

func (etcd *EtcdStore) DeleteVCluster(vclusterId string) error {
	return nil
}

func (etcd *EtcdStore) UpdateVCluster(*types.VCluster) error {
	return nil
}
