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
