package zk

import (
	"path"
	"sort"

	log "github.com/Sirupsen/logrus"
)

func (zk *ZKStore) GetLeader() string {
	p := "/leader-election"

	children, err := zk.list(p)
	if err != nil {
		log.Errorf("%v", err)
		return ""
	}

	sort.Strings(children)

	data, _, err := zk.get(path.Join(p, children[0]))
	if err != nil {
		log.Errorf("%v", err)
		return ""
	}

	return string(data)
}
