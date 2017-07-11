package zk

import (
	"fmt"
	"path"
	"strings"

	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
)

func (zk *ZKStore) CreateTask(aid string, task *types.Task) error {
	bs, err := encode(task)
	if err != nil {
		return err
	}

	p := path.Join(keyApp, aid, "tasks", task.ID)

	return zk.createAll(p, bs)
}

func (zk *ZKStore) UpdateTask(aid string, task *types.Task) error {
	bs, err := encode(task)
	if err != nil {
		return err
	}

	p := path.Join(keyApp, aid, "tasks", task.ID)

	return zk.set(p, bs)
}

func (zk *ZKStore) ListTaskHistory(aid, sid string) []*types.Task {
	return nil
}

func (zk *ZKStore) ListTasks(id string) ([]*types.Task, error) {
	p := path.Join(keyApp, id, "tasks")

	children, err := zk.list(p)
	if err != nil {
		log.Errorf("get app %s children(tasks) error: %v", id, err)
		return nil, err
	}

	tasks := make([]*types.Task, 0)
	for _, child := range children {
		p := path.Join(keyApp, id, "tasks", child)
		data, _, err := zk.get(p)
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

func (zk *ZKStore) DeleteTask(id string) error {
	p := join(id)

	return zk.del(p)
}

func (zk *ZKStore) GetTask(aid, tid string) (*types.Task, error) {
	p := path.Join(keyApp, aid, "tasks", tid)

	data, _, err := zk.get(p)
	if err != nil {
		if err == errNotExists {
			return nil, fmt.Errorf("task %s not exist", tid)
		}

		return nil, err
	}

	var task types.Task
	if err := decode(data, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

func join(id string) string {
	parts := strings.SplitN(id, ".", 3)

	return path.Join(keyApp, parts[2], "tasks", id)
}
