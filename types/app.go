package types

import (
	"sort"
	"strconv"
	"time"

	"github.com/Dataman-Cloud/swan/utils/fields"
	"github.com/Dataman-Cloud/swan/utils/labels"
)

const (
	OpStatusNoop     = "noop"
	OpStatusCreating = "creating"
	OpStatusScaling  = "scaling"
	OpStatusUpdating = "updating"
	OpStatusDeleting = "deleting"
	OpStatusRollback = "rollbacking"
)

type VersionList []*Version

func (vl VersionList) Len() int      { return len(vl) }
func (vl VersionList) Swap(i, j int) { vl[i], vl[j] = vl[j], vl[i] }
func (vl VersionList) Less(i, j int) bool {
	m, _ := strconv.Atoi(vl[i].ID)
	n, _ := strconv.Atoi(vl[j].ID)

	return m < n
}

func (vl VersionList) Sort() VersionList {
	sort.Sort(vl)

	return vl
}

func (vl VersionList) Reverse() VersionList {
	sort.Sort(sort.Reverse(vl))

	return vl
}

type Application struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Alias     string      `json:"alias"`
	RunAs     string      `json:"runAs"`
	Priority  int         `json:"priority"`
	Cluster   string      `json:"cluster"`
	OpStatus  string      `json:"operationStatus"`
	Tasks     int         `json:"tasks"`
	Version   []string    `json:"currentVersion"`
	Versions  VersionList `json:"versions"`
	Status    string      `json:"status"`
	Health    *Health     `json:"health"`
	CreatedAt time.Time   `json:"created"`
	UpdatedAt time.Time   `json:"updated"`
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
