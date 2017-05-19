package store

import "github.com/Sirupsen/logrus"

func (zk *ZKStore) UpdateFrameworkId(id string) error {
	return zk.createAll(keyFrameworkID, []byte(id))
}

func (zk *ZKStore) GetFrameworkId() string {
	bs, err := zk.get(keyFrameworkID)
	if err != nil {
		logrus.Errorln("zk GetFrameworkId error:", err)
	}
	return string(bs)
}
