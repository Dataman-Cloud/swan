package event

import (
	"encoding/json"

	"github.com/Dataman-Cloud/swan/utils/httpclient"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

const (
	//task_add and task_rm is used for dns/proxy service
	EventTypeTaskHealthy      = "task_healthy"
	EventTypeTaskWeightChange = "task_weight_change"
	EventTypeTaskUnhealthy    = "task_unhealthy"

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
	EventTypeAppStateFailed       = "app_state_failed"
	EventTypeAppStateUpdating     = "app_state_updating"
	EventTypeAppStateCancelUpdate = "app_state_cancel_update"
	EventTypeAppStateScaleUp      = "app_state_scale_up"
	EventTypeAppStateScaleDown    = "app_state_scale_down"
)

type Event struct {
	ID      string
	Type    string
	AppID   string
	AppMode string
	Payload interface{}
}

func (e *Event) String() string {
	bs, _ := json.Marshal(e)
	return string(bs)
}

func NewEvent(t string, payload interface{}) *Event {
	return &Event{
		ID:      uuid.NewV4().String(),
		Type:    t,
		Payload: payload,
	}
}

func SendEventByHttp(addr string, data interface{}) error {
	_, err := httpclient.NewDefaultClient().POST(context.TODO(), addr, nil, data, nil)

	return err
}
