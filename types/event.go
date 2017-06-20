package types

import (
	"encoding/json"
	"fmt"
)

const (
	EventTypeTaskHealthy      = "task_healthy"
	EventTypeTaskWeightChange = "task_weight_change"
	EventTypeTaskUnhealthy    = "task_unhealthy"
)

type TaskEvent struct {
	Type           string  `json:"type"`
	AppID          string  `json:"app_id"`
	AppAlias       string  `json:"app_alias"`
	VersionID      string  `json:"version_id"`
	AppVersion     string  `json:"app_version"`
	TaskID         string  `json:"task_id"`
	IP             string  `json:"task_ip"`
	Port           uint64  `json:"task_port"`
	Weight         float64 `json:"weihgt"`
	GatewayEnabled bool    `json:"gateway"`
}

type SSEEvent struct {
	name    string
	payload *TaskEvent
}

func (e *TaskEvent) Format() []byte {

	ev := &SSEEvent{
		name:    e.Type,
		payload: e,
	}

	bs, err := json.Marshal(ev)
	if err != nil {
		fmt.Printf("marshal event got error: %v", err)

		return []byte("")
	}

	return bs
	//e := &eventbus.Event{
	//	Type:    eventType,
	//	AppID:   slot.App.ID,
	//	AppMode: string(slot.App.Mode),
	//}

	//gatewayEnabled := true
	//if slot.Version.Gateway != nil {
	//	gatewayEnabled = slot.Version.Gateway.Enabled
	//}
	//payload := &types.TaskInfoEvent{
	//	TaskID:         slot.ID,
	//	AppID:          slot.App.ID,
	//	VersionID:      slot.Version.ID,
	//	AppVersion:     slot.Version.AppVersion,
	//	State:          slot.State,
	//	Healthy:        slot.healthy,
	//	ClusterID:      slot.App.ClusterID,
	//	RunAs:          slot.Version.RunAs,
	//	Weight:         slot.weight,
	//	AppName:        slot.App.Name,
	//	SlotIndex:      slot.Index,
	//	GatewayEnabled: gatewayEnabled,
	//}

	//if slot.App.IsFixed() {
	//	payload.IP = slot.Ip
	//	payload.Mode = string(APP_MODE_FIXED)
	//} else {
	//	payload.IP = slot.AgentHostName
	//	payload.Mode = string(APP_MODE_REPLICATES)
	//	if len(slot.CurrentTask.HostPorts) > 0 {
	//		payload.Port = uint32(slot.CurrentTask.HostPorts[0])
	//		payload.PortName = slot.Version.Container.Docker.PortMappings[0].Name
	//	}
	//}

	//e.Payload = payload

	//return e
}
