package scheduler

import (
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/scheduler/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddEvent(t *testing.T) {
	eventType := sched.Event_SUBSCRIBED
	event := &sched.Event{
		Type: sched.Event_UNKNOWN.Enum(),
	}

	s := NewScheduler("x.x.x.x:yyyy", nil, &mock.Store{}, "xxxx", nil, nil)
	s.AddEvent(eventType, event)

	e := <-s.GetEvent(eventType)
	assert.Equal(t, e.Type, sched.Event_UNKNOWN.Enum())
}
