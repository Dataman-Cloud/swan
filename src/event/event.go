package event

import (
	"github.com/satori/go.uuid"
)

const (
	//task_add and task_rm is used for dns/proxy service
	EventTypeTaskAdd = "task_add"
	EventTypeTaskRm  = "task_rm"

	EventTypeTaskStatePendingOffer   = "task_state_pending_offer"
	EventTypeTaskStatePendingKill    = "task_state_pending_killed"
	EventTypeTaskStateReap           = "task_state_pending_reap"
	EventTypeTaskStateStaging        = "task_state_staging"
	EventTypeTaskStateStarting       = "task_state_starting"
	EventTypeTaskStateRunning        = "task_state_running"
	EventTypeTaskStateKilling        = "task_state_killing"
	EventTypeTaskStateFinished       = "task_state_finished"
	EventTypeTaskStateFailed         = "task_state_failed"
	EventTypeTaskStateKilled         = "task_state_killed"
	EventTypeTaskStateError          = "task_state_error"
	EventTypeTaskStateLost           = "task_state_lost"
	EventTypeTaskStateDropped        = "task_state_dropped"
	EventTypeTaskStateUnreachable    = "task_state_unreachable"
	EventTypeTaskStateGone           = "task_state_gone"
	EventTypeTaskStateGoneByOperator = "task_state_gone_by_operator"
	EventTypeTaskStateUnknown        = "task_state_unknown"

	EventTypeAppStateCreating     = "app_state_creating"
	EventTypeAppStateDeletion     = "app_state_deletion"
	EventTypeAppStateNormal       = "app_state_normal"
	EventTypeAppStateUpdating     = "app_state_updating"
	EventTypeAppStateCancelUpdate = "app_state_cancel_update"
	EventTypeAppStateScaleUp      = "app_state_scale_up"
	EventTypeAppStateScaleDown    = "app_state_scale_down"
)

type Event struct {
	Id      string
	Type    string
	AppId   string
	Payload interface{}
}

func NewEvent(t string, payload interface{}) *Event {
	return &Event{
		Id:      uuid.NewV4().String(),
		Type:    t,
		Payload: payload,
	}
}

type TaskInfoEvent struct {
	Ip        string
	TaskId    string
	AppId     string
	Port      string
	State     string
	Healthy   bool
	ClusterId string
	RunAs     string
}

type AppInfoEvent struct {
	AppId     string
	Name      string
	State     string
	ClusterId string
	RunAs     string
}
