package inmemory

import (
	"errors"
	"sync"

	"github.com/Dataman-Cloud/swan/types"
)

var (
	ErrNotExists = errors.New("task does not exist")
)

type Registry struct {
	sync.RWMutex

	tasks map[string]*types.Task
}

func New() *Registry {
	return &Registry{
		tasks: make(map[string]*types.Task),
	}
}

func (r *Registry) Register(id string, task *types.Task) error {
	r.Lock()
	defer r.Unlock()

	r.tasks[id] = task

	return nil
}

func (r *Registry) Fetch(id string) (*types.Task, error) {
	r.RLock()
	defer r.RUnlock()

	t, exists := r.tasks[id]
	if !exists {
		return nil, ErrNotExists
	}

	return t, nil

}

func (r *Registry) Tasks() ([]*types.Task, error) {
	r.RLock()
	defer r.RUnlock()

	var (
		i   int
		out = make([]*types.Task, len(r.tasks))
	)

	for _, v := range r.tasks {
		out[i] = v
		i++
	}

	return out, nil
}

func (r *Registry) Update(id string, t *types.Task) error {
	r.Lock()
	defer r.Unlock()

	r.tasks[id] = t

	return nil
}

func (r *Registry) Delete(id string) error {
	r.Lock()
	defer r.Unlock()

	delete(r.tasks, id)

	return nil
}
