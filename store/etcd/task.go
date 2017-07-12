package etcd

import (
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/types"
)

func (s *EtcdStore) CreateTask(aid string, task *types.Task) error {
	bs, err := encode(task)
	if err != nil {
		return err
	}

	p := path.Join(keyApp, aid, keyTasks, task.ID)

	return s.create(p, bs)
}

func (s *EtcdStore) UpdateTask(aid string, task *types.Task) error {
	bs, err := encode(task)
	if err != nil {
		return err
	}

	p := path.Join(keyApp, aid, keyTasks, task.ID)

	return s.update(p, bs)
}

func (s *EtcdStore) ListTaskHistory(aid, sid string) []*types.Task {
	return nil
}

func (s *EtcdStore) ListTasks(id string) ([]*types.Task, error) {
	p := path.Join(keyApp, id, keyTasks)

	children, err := s.list(p)
	if err != nil {
		log.Errorf("get app %s children(tasks) error: %v", id, err)
		return nil, err
	}

	tasks := make([]*types.Task, 0)
	for child := range children {
		p := path.Join(keyApp, id, keyTasks, child)
		data, err := s.get(p)
		if err != nil {
			log.Errorf("get %s got error: %v", p, err)
			return nil, err
		}

		var t *types.Task
		if err := decode(data, &t); err != nil {
			log.Errorf("decode task %s got error: %v", id, err)
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (s *EtcdStore) DeleteTask(id string) error {
	var (
		parts = strings.SplitN(id, ".", 3)
		p     = path.Join(keyApp, parts[2], keyTasks, id)
	)

	return s.del(p, true)
}

func (s *EtcdStore) GetTask(aid, tid string) (*types.Task, error) {
	p := path.Join(keyApp, aid, keyTasks, tid)

	data, err := s.get(p)
	if err != nil {
		return nil, err
	}

	var task types.Task
	if err := decode(data, &task); err != nil {
		return nil, err
	}

	return &task, nil

}
