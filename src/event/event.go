package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/utils/httpclient"

	janitor "github.com/Dataman-Cloud/swan-janitor/src"
	"github.com/Dataman-Cloud/swan-resolver/nameserver"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

const (
	//task_add and task_rm is used for dns/proxy service
	EventTypeTaskHealthy   = "task_healthy"
	EventTypeTaskUnhealthy = "task_unhealthy"

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

func BuildResolverEvent(e *Event) (*nameserver.RecordGeneratorChangeEvent, error) {
	payload, ok := e.Payload.(*types.TaskInfoEvent)
	if !ok {
		return nil, errors.New("payload type error")
	}

	resolverEvent := &nameserver.RecordGeneratorChangeEvent{}
	if e.Type == EventTypeTaskHealthy {
		resolverEvent.Change = "add"
	} else {
		resolverEvent.Change = "del"
	}

	resolverEvent.Ip = payload.IP

	if payload.Mode == "replicates" {
		resolverEvent.Type = "srv"
		resolverEvent.Port = fmt.Sprintf("%d", payload.Port)
	} else {
		resolverEvent.Type = "a"
	}

	resolverEvent.DomainPrefix = strings.ToLower(strings.Replace(payload.TaskID, "-", ".", -1))

	return resolverEvent, nil
}

func BuildJanitorEvent(e *Event) (*janitor.TargetChangeEvent, error) {
	payload, ok := e.Payload.(*types.TaskInfoEvent)
	if !ok {
		return nil, errors.New("payload type error")
	}

	janitorEvent := &janitor.TargetChangeEvent{}
	if e.Type == EventTypeTaskHealthy {
		janitorEvent.Change = "add"
	} else {
		janitorEvent.Change = "del"
	}

	janitorEvent.TaskIP = payload.IP
	janitorEvent.TaskPort = payload.Port
	janitorEvent.AppID = payload.AppID
	janitorEvent.PortName = payload.PortName
	janitorEvent.TaskPort = payload.Port
	janitorEvent.TaskID = strings.ToLower(payload.TaskID)

	return janitorEvent, nil
}
