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

	// use task for compatability now, should be slot here
	Tasks    []*Task  `json:"tasks,omitempty"`
	Versions []string `json:"versions,omitempty"`
}

// use task for compatability now, should be slot here
// and together with task history
type Task struct {
	ID        string `json:"id,omitempty"`
	AppId     string `json:"appId,omitempty"`
	VersionId string `json:"versionId,omitempty"`

	Status string `json:"status,omitempty"`

	OfferId       string `json:"offerId,omitempty"`
	AgentId       string `json:"agentId,omitempty"`
	AgentHostname string `json:"agentHostname,omitempty"`

	History []*TaskHistory `json:"history,omitempty"`
}

type TaskHistory struct {
	ID        string `json:"id,omitempty"`
	AppId     string `json:"appId,omitempty"`
	VersionId string `json:"versionId,omitempty"`

	OfferId       string `json:"offerId,omitempty"`
	AgentId       string `json:"agentId,omitempty"`
	AgentHostname string `json:"agentHostname,omitempty"`

	State      string `json:"state,omitempty"`
	ExitReason string `json:"exitReason,omitempty"`
	Stdout     string `json:"stdout,omitempty"`
	Stderr     string `json:"stderr,omitempty"`
}
