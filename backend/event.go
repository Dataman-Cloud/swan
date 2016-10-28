package backend

import (
	"encoding/json"
	"net/http"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
)

func (b *Backend) EventStream(w http.ResponseWriter) error {
	var event *sched.Event

	notify := w.(http.CloseNotifier).CloseNotify()

	for {
		select {
		case event = <-b.sched.GetEvent(sched.Event_UPDATE):
			status := event.GetUpdate().GetStatus()
			ID := status.TaskId.GetValue()
			state := status.GetState()

			switch state {
			case mesos.TaskState_TASK_STAGING:
				WriteJSON(w, &types.Event{
					ID:      ID,
					Message: status.GetMessage(),
					Status:  "STAGING",
				})
			case mesos.TaskState_TASK_STARTING:
				WriteJSON(w, &types.Event{
					ID:      ID,
					Message: status.GetMessage(),
					Status:  "STARTING",
				})
			case mesos.TaskState_TASK_RUNNING:
				WriteJSON(w, &types.Event{
					ID:      ID,
					Message: status.GetMessage(),
					Status:  "RUNNING",
				})
			case mesos.TaskState_TASK_FINISHED:
				WriteJSON(w, &types.Event{
					ID:      ID,
					Message: status.GetMessage(),
					Status:  "FINISHED",
				})
			case mesos.TaskState_TASK_FAILED:
				WriteJSON(w, &types.Event{
					ID:      ID,
					Message: status.GetMessage(),
					Status:  "FAILED",
				})
			case mesos.TaskState_TASK_KILLED:
				WriteJSON(w, &types.Event{
					ID:      ID,
					Message: status.GetMessage(),
					Status:  "KILLED",
				})
			case mesos.TaskState_TASK_LOST:
				WriteJSON(w, &types.Event{
					ID:      ID,
					Message: status.GetMessage(),
					Status:  "LOST",
				})
			}
		case <-notify:
			logrus.Info("SSE Client closed")
			return nil
		}
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, message *types.Event) {
	f, _ := w.(http.Flusher)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	json.NewEncoder(w).Encode(message)
	f.Flush()
}
