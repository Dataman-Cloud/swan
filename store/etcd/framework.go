package etcd

import "math"

func (s *EtcdStore) UpdateFrameworkId(id string) error {
	return s.upsert(keyFrameworkID, []byte(id))
}

func (s *EtcdStore) GetFrameworkId() (string, int64) {
	bs, err := s.get(keyFrameworkID)
	if err != nil {
		return "", 0
	}
	return string(bs), math.MaxInt64 // FIXME never failover
}
