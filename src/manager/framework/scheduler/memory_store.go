package scheduler

import (
	"sync"

	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/utils/fields"
	"github.com/Dataman-Cloud/swan/src/utils/labels"
)

type AppFilterOptions struct {
	LabelSelectors []labels.Selector
	FieldSelector  fields.Selector
}

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

func (m *memoryStore) Filter(options AppFilterOptions) []*state.App {
	var apps []*state.App

	for _, app := range m.s {
		if !filterByLabelsSelectors(options.LabelSelectors, app.CurrentVersion.Labels) {
			continue
		}

		if !filterByFieldsSelectors(options.FieldSelector, app) {
			continue
		}

		apps = append(apps, app)
	}

	return apps
}

func filterByLabelsSelectors(labelsSelectors []labels.Selector, appLabels map[string]string) bool {
	for _, selector := range labelsSelectors {
		if !selector.Matches(labels.Set(appLabels)) {
			return false
		}
	}

	return true
}

func filterByFieldsSelectors(fieldSelector fields.Selector, app *state.App) bool {
	// TODO(upccup): there maybe exist better way to got a field/value map
	fieldMap := make(map[string]string)
	fieldMap["runAs"] = app.CurrentVersion.RunAs
	if !fieldSelector.Matches(fields.Set(fieldMap)) {
		return false
	}

	return true
}
