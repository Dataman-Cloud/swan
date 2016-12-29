package event

const (
	EVENT_TYPE_MESOS_SUBSCRIBED = "mesos_subscribed"
	EVENT_TYPE_MESOS_HEARTBEAT  = "mesos_heartbeat"
	EVENT_TYPE_MESOS_OFFERS     = "mesos_offers"
	EVENT_TYPE_MESOS_RESCIND    = "mesos_recind"
	EVENT_TYPE_MESOS_UPDATE     = "mesos_update"
	EVENT_TYPE_MESOS_FAILURE    = "mesos_failure"
	EVENT_TYPE_MESOS_MESSAGE    = "mesos_message"
	EVENT_TYPE_MESOS_ERROR      = "mesos_error"

	EVENT_TYPE_USER_INVALID_APPS = "user_invalidapps"
)

type Event interface {
	GetEventType() string
	GetEvent() interface{}
}
