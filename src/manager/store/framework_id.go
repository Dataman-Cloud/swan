package store

func (zk *ZKStore) UpdateFrameworkId(frameworkId string) error {
	op := &AtomicOp{
		Op:      OP_UPDATE,
		Entity:  ENTITY_FRAMEWORKID,
		Payload: frameworkId,
	}

	return zk.Apply(op, true)
}

func (zk *ZKStore) GetFrameworkId() string {
	zk.mu.RLock()
	defer zk.mu.RUnlock()

	return zk.Storage.FrameworkId
}
