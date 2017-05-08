package store

import "errors"

func (zk *ZkStore) CreateInstance(ins *Instance) error {
	op := &AtomicOp{
		Op:      OP_ADD,
		Entity:  ENTITY_INSTANCE,
		Param1:  ins.ID,
		Payload: ins,
	}

	return zk.Apply(op, true)
}

func (zk *ZkStore) DeleteInstance(idOrName string) error {
	i, _ := zk.GetInstance(idOrName)
	if i == nil {
		return nil
	}

	op := &AtomicOp{
		Op:     OP_REMOVE,
		Entity: ENTITY_INSTANCE,
		Param1: i.ID,
	}

	return zk.Apply(op, true)
}

func (zk *ZkStore) UpdateInstance(ins *Instance) error {
	op := &AtomicOp{
		Op:      OP_UPDATE,
		Entity:  ENTITY_INSTANCE,
		Param1:  ins.ID,
		Payload: ins,
	}

	return zk.Apply(op, true)
}

func (zk *ZkStore) GetInstance(idOrName string) (*Instance, error) {
	zk.mu.RLock()
	defer zk.mu.RUnlock()

	// by id
	ins, ok := zk.Storage.Instances[idOrName]
	if ok {
		return ins, nil
	}

	// by name
	var inss []*Instance
	inss, err := zk.ListInstances()
	if err != nil {
		return nil, err
	}
	for _, ins := range inss {
		if ins.Name == idOrName {
			return ins, nil
		}
	}

	return nil, errors.New("no such compose instance")
}

func (zk *ZkStore) ListInstances() ([]*Instance, error) {
	zk.mu.RLock()
	defer zk.mu.RUnlock()

	is := make([]*Instance, 0, 0)
	for _, i := range zk.Storage.Instances {
		is = append(is, i)
	}

	return is, nil
}
