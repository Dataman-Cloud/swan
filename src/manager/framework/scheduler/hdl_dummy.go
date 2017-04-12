package scheduler

import "github.com/Dataman-Cloud/swan/src/manager/framework/event"

func DummyHandler(s *Scheduler, ev event.Event) error {
	return nil
}
