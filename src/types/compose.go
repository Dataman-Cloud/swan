package types

import "time"

type ComposeRequest struct {
	Name      string                `json:"name"`
	Desc      string                `json:"desc"`
	YAMLRaw   string                `json:"yaml_raw"`
	YAMLEnv   map[string]string     `json:"yaml_env"`
	YAMLExtra map[string]*YamlExtra `json:"yaml_extra"`
}

type YamlExtra struct {
	Priority    uint              `json:"priority"`
	WaitDelay   uint              `json:"wait_delay"`
	PullAlways  bool              `json:"pull_always"`
	Resource    *Resource         `json:"resource"`
	Constraints string            `json:"constraints"`
	RunAs       string            `json:"runas"`
	URIs        []string          `json:"uris"`
	IPs         []string          `json:"ips"`
	Labels      map[string]string `json:"labels"` // extra labels: uid, username, vcluster ...
}

type Resource struct {
	CPU   float64  `json:"cpu"`
	Mem   float64  `json:"mem"`
	Disk  float64  `json:"disk"`
	Ports []uint64 `json:"ports"`
}

// TODO sigh ...
// almost same as store.Instance exclude `ServiceGroup`
type ComposeInstance struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Desc         string                 `json:"desc"`
	Status       string                 `json:"status"`
	ErrMsg       string                 `json:"errmsg"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	ServiceGroup map[string]interface{} `json:"service_group"` // just for output
	YAMLRaw      string                 `json:"yaml_raw"`
	YAMLEnv      map[string]string      `json:"yaml_env"`
	YAMLExtra    map[string]*YamlExtra  `json:"yaml_extra"`
}

// same as api.instanceWraper
type ComposeInstanceWrapper struct {
	*ComposeInstance
	Apps []*App `json:"apps"`
}
