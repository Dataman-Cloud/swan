package types

import (
	"encoding/json"
	"fmt"

	"github.com/Dataman-Cloud/swan/agent/janitor/upstream"
	"github.com/Dataman-Cloud/swan/agent/resolver"
)

const (
	EventTypeTaskHealthy      = "task_healthy"
	EventTypeTaskWeightChange = "task_weight_change"
	EventTypeTaskUnhealthy    = "task_unhealthy"
)

type CombinedEvents struct {
	Event *TaskEvent
	Proxy *upstream.BackendCombined // built from event
	DNS   *resolver.Record          // built from event
}

type TaskEvent struct {
	Type           string  `json:"type"`
	AppID          string  `json:"app_id"`
	AppAlias       string  `json:"app_alias"`  // for proxy
	AppListen      string  `json:"app_listen"` // for proxy
	AppSticky      bool    `json:"app_sticky"` // for proxy
	VersionID      string  `json:"version_id"`
	AppVersion     string  `json:"app_version"`
	TaskID         string  `json:"task_id"`
	IP             string  `json:"task_ip"`
	Port           uint64  `json:"task_port"`
	TargetPort     uint64  `json:"target_port"`
	Weight         float64 `json:"weihgt"`
	GatewayEnabled bool    `json:"gateway"` // for proxy
}

// Format format task events to SSE text
func (e *TaskEvent) Format() []byte {
	bs, _ := json.Marshal(e)
	return []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", e.Type, string(bs)))
}
