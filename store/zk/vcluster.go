package zk

import (
	"path"

	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
)

func (zk *ZKStore) ListVClusters() ([]*types.VCluster, error) {
	nodes, err := zk.list(keyVCluster)
	if err != nil {
		log.Errorln("zk ListVClusters error:", err)
		return nil, err
	}

	vclusters := []*types.VCluster{}
	for _, node := range nodes {
		vc, err := zk.GetVCluster(node)
		if err != nil {
			log.Errorf("get vcluster error: %v", err)
			continue
		}

		vclusters = append(vclusters, vc)
	}

	return vclusters, nil
}

func (zk *ZKStore) GetVCluster(clusterId string) (*types.VCluster, error) {
	p := path.Join(keyVCluster, clusterId)

	data, _, err := zk.get(p)
	if err != nil {
		return nil, err
	}

	vcluster := new(types.VCluster)
	if err := decode(data, vcluster); err != nil {
		return nil, err
	}

	nodes, err := zk.getNodes(p, clusterId)
	if err != nil {
		return nil, err
	}

	vcluster.Nodes = nodes

	return vcluster, nil
}

func (zk *ZKStore) CreateVCluster(vcluster *types.VCluster) error {
	p := path.Join(keyVCluster, vcluster.ID)

	bs, err := encode(vcluster)
	if err != nil {
		return err
	}

	return zk.create(p, bs)
}

func (zk *ZKStore) VClusterExists(name string) bool {
	vclusters, _ := zk.ListVClusters()

	for _, vcluster := range vclusters {
		if vcluster.Name == name {
			return true
		}
	}

	return false
}

func (zk *ZKStore) DeleteVCluster(vclusterId string) error {
	p := path.Join(keyVCluster, vclusterId)

	if err := zk.del(p); err != nil {
		log.Errorf("delete vcluster %s got error: %v", err)
		return err
	}

	return nil
}

func (zk *ZKStore) UpdateVCluster(vcluster *types.VCluster) error {
	bs, err := encode(vcluster)
	if err != nil {
		return err
	}

	p := path.Join(keyVCluster, vcluster.ID)

	return zk.set(p, bs)
}

func (zk *ZKStore) getNodes(p, vclusterId string) ([]*types.Node, error) {
	children, err := zk.list(path.Join(p, "nodes"))
	if err != nil {
		return nil, err
	}

	nodes := make([]*types.Node, 0)
	for _, child := range children {
		p := path.Join(keyVCluster, vclusterId, "nodes", child)
		data, _, err := zk.get(p)
		if err != nil {
			continue
		}

		node := new(types.Node)
		if err := decode(data, node); err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
