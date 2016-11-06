package scheduler

import (
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/scheduler/mock"
	"github.com/golang/protobuf/proto"
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

func TestSchedulerStart(t *testing.T) {
	fw := &mesos.FrameworkInfo{
		User:            proto.String("testuser"),
		Name:            proto.String("swan"),
		Hostname:        proto.String("x.x.x.x"),
		FailoverTimeout: proto.Float64(5),
		Id: &mesos.FrameworkID{
			Value: proto.String("xxxx-yyyy-zzzz"),
		},
	}

	s := NewScheduler("x.x.x.x:yyyy", fw, &mock.Store{}, "xxxx", nil, nil)
	s.Start()
}
