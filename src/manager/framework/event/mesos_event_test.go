package event

import (
	"testing"

	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
	"github.com/stretchr/testify/assert"
)

func TestMesosEventGetEventTypeSubscribed(t *testing.T) {
	me := &MesosEvent{
		EventType: sched.Event_SUBSCRIBED,
		Event:     &sched.Event{},
	}

	assert.Equal(t, me.GetEventType(), EVENT_TYPE_MESOS_SUBSCRIBED)
}

func TestMesosEventGetEventTypeHeartbeat(t *testing.T) {
	me := &MesosEvent{
		EventType: sched.Event_HEARTBEAT,
		Event:     &sched.Event{},
	}

	assert.Equal(t, me.GetEventType(), EVENT_TYPE_MESOS_HEARTBEAT)
}

func TestMesosEventGetEventTypeOffers(t *testing.T) {
	me := &MesosEvent{
		EventType: sched.Event_OFFERS,
		Event:     &sched.Event{},
	}

	assert.Equal(t, me.GetEventType(), EVENT_TYPE_MESOS_OFFERS)
}

func TestMesosEventGetEventTypeRESCIND(t *testing.T) {
	me := &MesosEvent{
		EventType: sched.Event_RESCIND,
		Event:     &sched.Event{},
	}

	assert.Equal(t, me.GetEventType(), EVENT_TYPE_MESOS_RESCIND)
}

func TestMesosEventGetEventTypeUpdate(t *testing.T) {
	me := &MesosEvent{
		EventType: sched.Event_UPDATE,
		Event:     &sched.Event{},
	}

	assert.Equal(t, me.GetEventType(), EVENT_TYPE_MESOS_UPDATE)
}

func TestMesosEventGetEventTypeFailure(t *testing.T) {
	me := &MesosEvent{
		EventType: sched.Event_FAILURE,
		Event:     &sched.Event{},
	}

	assert.Equal(t, me.GetEventType(), EVENT_TYPE_MESOS_FAILURE)
}

func TestMesosEventGetEventTypeMessage(t *testing.T) {
	me := &MesosEvent{
		EventType: sched.Event_MESSAGE,
		Event:     &sched.Event{},
	}

	assert.Equal(t, me.GetEventType(), EVENT_TYPE_MESOS_MESSAGE)
}

func TestMesosEventGetEventTypeError(t *testing.T) {
	me := &MesosEvent{
		EventType: sched.Event_ERROR,
		Event:     &sched.Event{},
	}

	assert.Equal(t, me.GetEventType(), EVENT_TYPE_MESOS_ERROR)
}

func TestMesosEventGetEvent(t *testing.T) {
	e := &sched.Event{}
	me := &MesosEvent{
		EventType: sched.Event_ERROR,
		Event:     e,
	}

	assert.Equal(t, me.GetEvent(), e)
}
