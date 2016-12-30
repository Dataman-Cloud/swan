package event

import (
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
)

type MesosEvent struct {
	EventType sched.Event_Type
	Event     *sched.Event
}

func (me *MesosEvent) GetEventType() string {
	switch me.EventType {
	case sched.Event_SUBSCRIBED:
		return EVENT_TYPE_MESOS_SUBSCRIBED
	case sched.Event_HEARTBEAT:
		return EVENT_TYPE_MESOS_HEARTBEAT
	case sched.Event_OFFERS:
		return EVENT_TYPE_MESOS_OFFERS
	case sched.Event_RESCIND:
		return EVENT_TYPE_MESOS_RESCIND
	case sched.Event_UPDATE:
		return EVENT_TYPE_MESOS_UPDATE
	case sched.Event_FAILURE:
		return EVENT_TYPE_MESOS_FAILURE
	case sched.Event_MESSAGE:
		return EVENT_TYPE_MESOS_MESSAGE
	case sched.Event_ERROR:
		return EVENT_TYPE_MESOS_ERROR
	default:
		panic("not known event type")
	}
}

func (me *MesosEvent) GetEvent() interface{} {
	return me.Event
}
