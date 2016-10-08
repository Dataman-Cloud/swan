package scheduler

import "github.com/Dataman-Cloud/swan/types"

type Registry interface {
	// Register a new task by ID in the memory registry
	Register(string, *types.Task) error

	// Tasks returns all tasks in the registry
	Tasks() ([]*types.Task, error)

	// Delete removes the task by ID from the registry
	Delete(string) error

	// Fetch returns a specific task in the registry by ID
	Fetch(string) (*types.Task, error)

	// Update finds the task in the registry for the ID and updates it's data
	Update(string, *types.Task) error
}
