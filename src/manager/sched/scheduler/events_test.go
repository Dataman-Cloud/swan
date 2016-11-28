package scheduler

import (
	"github.com/Dataman-Cloud/swan/manager/sched/mock"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddEvent(t *testing.T) {
	eventType := sched.Event_SUBSCRIBED
	event := &sched.Event{
		Type: sched.Event_SUBSCRIBED.Enum(),
	}

	s := NewScheduler(FakeConfig(), &mock.Store{})
	s.AddEvent(eventType, event)

	e := <-s.GetEvent(eventType)
	assert.Equal(t, e.Type, sched.Event_SUBSCRIBED.Enum())

	evt := sched.Event_UNKNOWN
	ev := &sched.Event{
		Type: sched.Event_UNKNOWN.Enum(),
	}

	err := s.AddEvent(evt, ev)
	assert.NotNil(t, err)
}

func TestGetEvent(t *testing.T) {
	s := NewScheduler(FakeConfig(), &mock.Store{})
	ev := s.GetEvent(sched.Event_UNKNOWN)
	assert.Nil(t, ev)
}
