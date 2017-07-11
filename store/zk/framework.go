package zk

import log "github.com/Sirupsen/logrus"

func (zk *ZKStore) UpdateFrameworkId(id string) error {
	return zk.set(keyFrameworkID, []byte(id))
}

func (zk *ZKStore) GetFrameworkId() (string, int64) {
	bs, stat, err := zk.get(keyFrameworkID)
	if err != nil {
		log.Errorln("zk GetFrameworkId error:", err)
		return "", 0
	}

	return string(bs), stat.Mtime
}
