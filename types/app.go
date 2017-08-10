package types

import (
	"time"

	"github.com/Dataman-Cloud/swan/utils/fields"
	"github.com/Dataman-Cloud/swan/utils/labels"
)

const (
	OpStatusNoop             = "noop"
	OpStatusCreating         = "creating"
	OpStatusScalingUp        = "scaling_up"
	OpStatusScalingDown      = "scaling_down"
	OpStatusUpdating         = "updating"
	OpStatusCanaryUpdating   = "canary_updating"
	OpStatusCanaryUnfinished = "canary_unfinished"
	OpStatusWeightUpdating   = "weight_updating"
	OpStatusStarting         = "starting"
	OpStatusStopping         = "stopping"
	OpStatusDeleting         = "deleting"
	OpStatusRollback         = "rollbacking"
)

type Application struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	RunAs           string          `json:"runAs"`
	Cluster         string          `json:"cluster"`
	OpStatus        string          `json:"operationStatus"`
	Progress        int             `json:"progress"`         // only used within app update
	ProgressDetails map[string]bool `json:"progress_details"` // only used within app update, taskid -> updated
	TaskCount       int             `json:"task_count"`
	Version         []string        `json:"currentVersion"`
	VersionCount    int             `json:"version_count"`
	Status          string          `json:"status"`
	TasksStatus     map[string]int  `json:"tasks_status"`
	Health          *Health         `json:"health"`
	ErrMsg          string          `json:"errmsg"`
	CreatedAt       time.Time       `json:"created"`
	UpdatedAt       time.Time       `json:"updated"`
}

type AppFilterOptions struct {
	LabelsSelector labels.Selector
	FieldsSelector fields.Selector
}

type Health struct {
	Total     int64 `json:"total"`
	Healthy   int64 `json:"healthy"`
	UnHealthy int64 `json:"unhealthy"`
	UnSet     int64 `json:"unset"`
}
