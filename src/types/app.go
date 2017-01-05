package types

import (
	"time"
)

type App struct {
	ID               string    `json:"id,omitempty"`
	Name             string    `json:"name,omitempty"`
	Instances        int       `json:"instances,omitempty"`
	UpdatedInstances int       `json:"updatedInstances,omitempty"`
	RunningInstances int       `json:"runningInstances"`
	RunAs            string    `json:"runAs,omitempty"`
	Priority         int       `json:"priority"`
	ClusterID        string    `json:"clusterID,omitempty"`
	Status           string    `json:"status,omitempty"`
	Created          time.Time `json:"created,omitempty"`
	Updated          time.Time `json:"updated,omitempty"`
	Mode             string    `json:"mode,omitempty"`
	State            string    `json:"state"`

	// use task for compatability now, should be slot here
	Tasks          []*Task  `json:"tasks,omitempty"`
	CurrentVersion *Version `json:"currentVersion"`
	// use when app updated, ProposedVersion can either be commit or revert
	ProposedVersion *Version `json:"proposedVersion,omitempty"`
	Versions        []string `json:"versions,omitempty"`
	IP              []string `json:"ip,omitempty"`

	// current version related info
	Labels      map[string]string `json:"labels,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Constraints []string          `json:"constraints,omitempty"`
	URIs        []string          `json:"uris,omitempty"`
}

// use task for compatability now, should be slot here
// and together with task history
type Task struct {
	ID          string       `json:"id,omitempty"`
	AppID       string       `json:"appId,omitempty"`
	VersionID   string       `json:"versionId,omitempty"`
	CurrentTask *TaskHistory `json:"currentTask,omitempty"`

	Status string `json:"status"`

	OfferID       string `json:"offerID,omitempty"`
	AgentID       string `json:"agentID,omitempty"`
	AgentHostname string `json:"agentHostname,omitempty"`

	CPU  float64 `json:"cpu,omitempty"`
	Mem  float64 `json:"mem,omitempty"`
	Disk float64 `json:"disk,omitempty"`

	History []*TaskHistory `json:"history,omitempty"`

	IP string `json:"ip,omitempty"`

	Created time.Time `json:"created,omitempty"`

	Image   string `json:"image,omitempty"`
	Healthy bool   `json:"healthy"`
}

type TaskHistory struct {
	ID        string `json:"id,omitempty"`
	AppID     string `json:"appID,omitempty"`
	VersionID string `json:"versionID,omitempty"`

	OfferID       string `json:"offerID,omitempty"`
	AgentID       string `json:"agentID,omitempty"`
	AgentHostname string `json:"agentHostname,omitempty"`

	CPU  float64 `json:"cpu,omitempty"`
	Mem  float64 `json:"mem,omitempty"`
	Disk float64 `json:"disk,omitempty"`

	State  string `json:"state,omitempty"`
	Reason string `json:"reason,omitempty"`
	Stdout string `json:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty"`
}

type Stats struct {
	ClusterID string `json:"clusterID"`

	AppCount  int `json:"appCount"`
	TaskCount int `json:"taskCount"`

	CpuTotalOffered  float64 `json:"cpuTotalOffered"`
	MemTotalOffered  float64 `json:"memTotalOffered"`
	DiskTotalOffered float64 `json:"diskTotalOffered"`

	CpuTotalUsed  float64 `json:"cpuTotalUsed"`
	MemTotalUsed  float64 `json:"memTotalUsed"`
	DiskTotalUsed float64 `json:"diskTotalUsed"`

	AppStats map[string]int `json:"appStats,omitempty"`
}

type ProceedUpdateParam struct {
	Instances int `json:"instances"`
}

type ScaleUpParam struct {
	Instances int      `json:"instances"`
	IPs       []string `json:"ips"`
}

type ScaleDownParam struct {
	Instances int `json:"instances"`
}
