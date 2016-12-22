package scheduler

import (
	"sync"

	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/utils/fields"
	"github.com/Dataman-Cloud/swan/src/utils/labels"
)

type AppFilterOptions struct {
	LabelsSelector labels.Selector
	FieldsSelector fields.Selector
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
		if !filterByLabelsSelectors(options.LabelsSelector, app.CurrentVersion.Labels) {
			continue
		}

		if !filterByFieldsSelectors(options.FieldsSelector, app) {
			continue
		}

		apps = append(apps, app)
	}

	return apps
}

func filterByLabelsSelectors(labelsSelector labels.Selector, appLabels map[string]string) bool {
	if labelsSelector == nil {
		return true
	}

	return labelsSelector.Matches(labels.Set(appLabels))
}

func filterByFieldsSelectors(fieldSelector fields.Selector, app *state.App) bool {
	if fieldSelector == nil {
		return true
	}

	// TODO(upccup): there maybe exist better way to got a field/value map
	fieldMap := make(map[string]string)
	fieldMap["runAs"] = app.CurrentVersion.RunAs
	return fieldSelector.Matches(fields.Set(fieldMap))
}
