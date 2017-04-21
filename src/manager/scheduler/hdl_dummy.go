package scheduler

import "github.com/Dataman-Cloud/swan/src/manager/event"

func DummyHandler(s *Scheduler, ev event.Event) error {
	return nil
}
