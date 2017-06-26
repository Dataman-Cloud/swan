package zk

import (
	"path"
	"sort"

	log "github.com/Sirupsen/logrus"
)

const LeaderElectionPath = "/leader-election" // TODO(nmg)

func (zk *ZKStore) GetLeader() (string, error) {
	children, err := zk.list(LeaderElectionPath)
	if err != nil {
		log.Errorf("%v", err)
		return "", err
	}

	sort.Strings(children)

	data, _, err := zk.get(path.Join(LeaderElectionPath, children[0]))
	if err != nil {
		log.Errorf("%v", err)
		return "", err
	}

	return string(data), nil
}
