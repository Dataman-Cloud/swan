package scheduler

import (
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/scheduler/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchedulerSend(t *testing.T) {
	s := NewScheduler("x.x.x.x:yyyy", nil, &mock.Store{}, "xxxx", nil, nil)
	call := &sched.Call{
		Type: sched.Call_SUBSCRIBE.Enum(),
		Subscribe: &sched.Call_Subscribe{
			FrameworkInfo: nil,
		},
	}

	_, err := s.send(call)
	assert.NotNil(t, err)
}

func TestSchedulerStop(t *testing.T) {
	s := NewScheduler("x.x.x.x:yyyy", nil, &mock.Store{}, "xxxx", nil, nil)
	s.stop()
}
