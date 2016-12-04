package api

import (
	"time"
)

type App struct {
	ID                string    `json:"id,omitempty"`
	Name              string    `json:"name,omitempty"`
	Instances         int       `json:"instances,omitempty"`
	UpdatedInstances  int       `json:"updatedInstances,omitempty"`
	RunningInstances  int       `json:"runningInstances,omitempty"`
	RollbackInstances int       `json:"rollbackInstances,omitempty"`
	RunAs             string    `json:"runAs,omitempty"`
	ClusterId         string    `json:"clusterId,omitempty"`
	Status            string    `json:"status,omitempty"`
	Created           time.Time `json:"created,omitempty"`
	Updated           time.Time `json:"updated,omitempty"`
	Mode              string    `json:"mode,omitempty"`
}
