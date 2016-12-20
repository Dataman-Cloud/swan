package scheduler

import (
	"sync"

	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
)

// memoryStore implements a Store in memory.
type memoryStore struct {
	s map[string]*state.App
	sync.RWMutex
}

// NewMemoryStore initializes a new memory store.
func NewMemoryStore() *memoryStore {
	return &memoryStore{
		s: make(map[string]*state.App),
	}
}

// Add appends a new app to the memory store.
// It overrides the id if it existed before.
func (m *memoryStore) Add(id string, app *state.App) {
	m.Lock()
	m.s[id] = app
	m.Unlock()
}

// Get returns an app from the store by id
func (m *memoryStore) Get(id string) *state.App {
	var res *state.App
	m.Lock()
	res = m.s[id]
	m.Unlock()
	return res
}

// Delete removes an app from the state by id
func (m *memoryStore) Delete(id string) {
	m.Lock()
	delete(m.s, id)
	m.Unlock()
}

func (m *memoryStore) Data() map[string]*state.App {
	return m.s
}
