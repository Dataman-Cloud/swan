package store

func (zk *ZkStore) UpdateFrameworkId(frameworkId string) error {
	op := &StoreOp{
		Op:      OP_UPDATE,
		Entity:  ENTITY_FRAMEWORKID,
		Payload: frameworkId,
	}

	return zk.Apply(op)
}

func (zk *ZkStore) GetFrameworkId() string {
	return zk.FrameworkId
}
