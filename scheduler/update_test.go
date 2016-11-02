package scheduler

import (
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/scheduler/mock"
	"github.com/golang/protobuf/proto"
	"testing"
)

func TestStatus(t *testing.T) {
	s := NewScheduler("x.x.x.x:yyyy", nil, &mock.Store{}, "xxxxx", nil, nil)
	ev := &sched.Event{
		Type: sched.Event_UPDATE.Enum(),
		Update: &sched.Event_Update{
			Status: &mesos.TaskStatus{
				TaskId: &mesos.TaskID{
					Value: proto.String("xxxxxxxx"),
				},
				State:   mesos.TaskState_TASK_RUNNING.Enum(),
				Message: proto.String("zzzzzzz"),
				Uuid:    []byte(`xxxxxxxxxx`),
				AgentId: &mesos.AgentID{
					Value: proto.String("yyyyyyyyy"),
				},
			},
		},
	}
	status := ev.GetUpdate().GetStatus()
	s.status(status)
}
