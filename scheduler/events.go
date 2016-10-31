package scheduler

import (
	"fmt"

	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Sirupsen/logrus"
)

type Events map[sched.Event_Type]chan *sched.Event

func (s *Scheduler) AddEvent(eventType sched.Event_Type, event *sched.Event) error {
	logrus.WithFields(logrus.Fields{"type": eventType}).Debug("Received event from master.")
	if c, ok := s.events[eventType]; ok {
		c <- event
		return nil
	}
	return fmt.Errorf("unknown event type: %v", eventType)
}

func (s *Scheduler) GetEvent(eventType sched.Event_Type) chan *sched.Event {
	if c, ok := s.events[eventType]; ok {
		return c
	} else {
		return nil
	}
}
