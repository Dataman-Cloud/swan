package event

import (
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
)

type MesosEvent struct {
	EventType sched.Event_Type
	Event     *sched.Event
}
