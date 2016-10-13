package scheduler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/types"
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

func (s *Scheduler) EventStream(w http.ResponseWriter) error {
	var event *sched.Event

	notify := w.(http.CloseNotifier).CloseNotify()

	for {
		select {
		case event = <-s.GetEvent(sched.Event_SUBSCRIBED):
			WriteJSON(w, fmt.Sprintf("Subscription successful with FrameworkID=%s", event.GetSubscribed().FrameworkId.GetValue()))
		case event = <-s.GetEvent(sched.Event_OFFERS):
			if event != nil {
				WriteJSON(w, fmt.Sprintf("Received %d offer(s).", len(event.Offers.Offers)))
			}
		case event = <-s.GetEvent(sched.Event_RESCIND):
		case event = <-s.GetEvent(sched.Event_UPDATE):
			status := event.GetUpdate().GetStatus()
			state := status.GetState()
			switch state {
			case mesos.TaskState_TASK_STAGING:
				WriteJSON(w, fmt.Sprintf("Task %s is STAGING", status.TaskId.GetValue()))
			case mesos.TaskState_TASK_STARTING:
				WriteJSON(w, fmt.Sprintf("Task %s is STARTING", status.TaskId.GetValue()))
			case mesos.TaskState_TASK_RUNNING:
				WriteJSON(w, fmt.Sprintf("Task %s is RUNNING", status.TaskId.GetValue()))
			case mesos.TaskState_TASK_FINISHED:
				WriteJSON(w, fmt.Sprintf("Task %s is FINISHED", status.TaskId.GetValue()))
			case mesos.TaskState_TASK_FAILED:
				WriteJSON(w, fmt.Sprintf("Task %s is FAILED", status.TaskId.GetValue()))
			case mesos.TaskState_TASK_KILLED:
				WriteJSON(w, fmt.Sprintf("Task %s is KILLED", status.TaskId.GetValue()))
			case mesos.TaskState_TASK_LOST:
				WriteJSON(w, fmt.Sprintf("Task %s is LOST", status.TaskId.GetValue()))
			}
		case event = <-s.GetEvent(sched.Event_MESSAGE):
		case event = <-s.GetEvent(sched.Event_FAILURE):
		case event = <-s.GetEvent(sched.Event_ERROR):
		case event = <-s.GetEvent(sched.Event_HEARTBEAT):
		case <-notify:
			logrus.Info("SSE Client closed")
			return nil
		}
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, message string) {
	f, _ := w.(http.Flusher)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	json.NewEncoder(w).Encode(&types.Event{
		Message: message,
	})
	f.Flush()
}
