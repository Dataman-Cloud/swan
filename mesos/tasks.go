package mesos

import (
	//"fmt"
	"sync"

	"github.com/Dataman-Cloud/swan/mesosproto"
	"github.com/golang/protobuf/proto"
)

type Tasks struct {
	sync.RWMutex
	tasks []*Task
}

func NewTasks() *Tasks {
	return &Tasks{}
}

func (t *Tasks) Build(offer *Offer) {
	port := offer.Ports()

	t.RLock()
	defer t.RUnlock()

	for _, task := range t.tasks {
		if p := port(); p != 0 {
			task.cfg.Port = p
		}

		task.AgentId = &mesosproto.AgentID{
			Value: proto.String(offer.GetAgentId()),
		}

		task.Build()
	}
}

func (t *Tasks) GetName() string {
	t.RLock()
	defer t.RUnlock()

	return t.tasks[0].GetName()
}

func (t *Tasks) taskInfos() (tasks []*mesosproto.TaskInfo) {
	t.RLock()
	defer t.RUnlock()

	for _, t := range t.tasks {
		tasks = append(tasks, &t.TaskInfo)
	}

	return
}

func (t *Tasks) Push(task *Task) {
	t.push(task)
}

func (t *Tasks) push(task *Task) {
	t.Lock()
	defer t.Unlock()

	t.tasks = append(t.tasks, task)
}

func (t *Tasks) Len() int {
	t.RLock()
	defer t.RUnlock()

	return len(t.tasks)
}

func (t *Tasks) Tasks() []*Task {
	t.RLock()
	defer t.RUnlock()

	return t.tasks
}
