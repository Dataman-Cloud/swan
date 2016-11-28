package scheduler

import (
	"github.com/Dataman-Cloud/swan/manager/apiserver"
	"github.com/Dataman-Cloud/swan/manager/sched/mock"
	"github.com/Dataman-Cloud/swan/manager/swancontext"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func FakeConfig() util.Scheduler {
	return util.Scheduler{
		MesosMasters:       []string{"xx.x.x.x:yyyy"},
		MesosFrameworkUser: "root",
		Hostname:           "foobar",
	}
}

func FakeSwanContext() *swancontext.SwanContext {
	return &swancontext.SwanContext{
		Store:     &mock.Store{},
		ApiServer: apiserver.NewApiServer("x", "x"),
	}
}

func TestSchedulerSend(t *testing.T) {
	s := NewScheduler(FakeConfig(), &mock.Store{})
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
	s := NewScheduler(FakeConfig(), &mock.Store{})
	s.stop()
}

func TestSchedulerStart(t *testing.T) {
	s := NewScheduler(FakeConfig(), &mock.Store{})
	s.Run()
}
