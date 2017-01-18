package swan

import (
	"errors"
	"fmt"

	"github.com/Dataman-Cloud/swan/src/types"
)

// EventsChannel is a channel to receive events upon
type EventsChannel chan *Event

type Event struct {
	ID    string
	Event string
	Data  interface{}
}

const (
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

func GetEvent(eventType string) (*Event, error) {
	event := new(Event)
	switch eventType {
	case EventTypeTaskStatePendingOffer,
		EventTypeTaskStatePendingKill,
		EventTypeTaskStateReap,
		EventTypeTaskStateStaging,
		EventTypeTaskStateStarting,
		EventTypeTaskStateRunning,
		EventTypeTaskStateKilling,
		EventTypeTaskStateFinished,
		EventTypeTaskStateFailed,
		EventTypeTaskStateKilled,
		EventTypeTaskStateError,
		EventTypeTaskStateLost,
		EventTypeTaskStateDropped,
		EventTypeTaskStateUnreachable,
		EventTypeTaskStateGone,
		EventTypeTaskStateGoneByOperator,
		EventTypeTaskStateUnknown:
		event.Data = new(types.TaskInfoEvent)
	case EventTypeAppStateCreating,
		EventTypeAppStateDeletion,
		EventTypeAppStateNormal,
		EventTypeAppStateUpdating,
		EventTypeAppStateCancelUpdate,
		EventTypeAppStateScaleUp,
		EventTypeAppStateScaleDown:
		event.Data = new(types.AppInfoEvent)
	default:
		return nil, errors.New(fmt.Sprintf("The event type %s was not found or supported", eventType))
	}
	return event, nil
}
