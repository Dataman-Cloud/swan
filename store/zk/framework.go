package zk

import (
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/mesos"
	log "github.com/Sirupsen/logrus"
)

func (zk *ZKStore) UpdateFrameworkId(id string) error {
	return zk.set(keyFrameworkID, []byte(id))
}

func (zk *ZKStore) GetFrameworkId() string {
	bs, stat, err := zk.get(keyFrameworkID)
	if err != nil {
		if !strings.Contains(err.Error(), "node does not exist") {
			log.Errorln("zk GetFrameworkId error:", err)
		}

		return ""
	}

	if time.Now().Unix()-(stat.Mtime/1000) >= mesos.DefaultFrameworkFailoverTimeout {
		log.Debugln("framework failover time exceed")

		return ""
	}

	return string(bs)
}
