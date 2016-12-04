package engine

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"

	"github.com/Sirupsen/logrus"
)

func UpdateHandler(h *Handler) (*Handler, error) {
	logrus.WithFields(logrus.Fields{"handler": "update"}).Debugf("logger handler report got event type: %s", h.MesosEvent.EventType)

	taskStatus := h.MesosEvent.Event.GetUpdate().GetStatus()
	AckUpdateEvent(h, taskStatus)

	slotName := taskStatus.TaskId.GetValue()
	taskState := taskStatus.GetState()
	//healthy := taskStatus.GetHealthy()
	//fmt.Println(healthy)
	//fmt.Println(state)

	slotIndex_, AppId := strings.Split(slotName, "-")[0], strings.Split(slotName, "-")[1]

	slotIndex, _ := strconv.ParseInt(slotIndex_, 10, 32)

	fmt.Println(AppId)
	fmt.Println(slotIndex)
	fmt.Println(h.EngineRef.Apps)

	switch taskState {
	case mesos.TaskState_TASK_STAGING:
	case mesos.TaskState_TASK_STARTING:
	case mesos.TaskState_TASK_RUNNING:
		h.EngineRef.Apps[AppId].Slots[slotIndex].SetState(state.SLOT_STATE_TASK_RUNNING)

	case mesos.TaskState_TASK_FINISHED:
		logrus.Infof("Task Finished, message: %s", taskStatus.GetMessage())
		h.EngineRef.Apps[AppId].Slots[slotIndex].SetState(state.SLOT_STATE_TASK_FINISHED)

	case mesos.TaskState_TASK_FAILED:
		logrus.Infof("Task Failed, message: %s", taskStatus.GetMessage())
		h.EngineRef.Apps[AppId].Slots[slotIndex].SetState(state.SLOT_STATE_TASK_FAILED)

	case mesos.TaskState_TASK_KILLED:
		logrus.Infof("Task Killed, message: %s", taskStatus.GetMessage())
		h.EngineRef.Apps[AppId].Slots[slotIndex].SetState(state.SLOT_STATE_TASK_KILLED)

	case mesos.TaskState_TASK_LOST:
		logrus.Infof("Task Lost, message: %s", taskStatus.GetMessage())
	}

	h.EngineRef.InvalidateApps()

	return h, nil
}

func AckUpdateEvent(h *Handler, taskStatus *mesos.TaskStatus) {
	if taskStatus.GetUuid() != nil {
		call := &sched.Call{
			FrameworkId: h.EngineRef.Scheduler.Framework.GetId(),
			Type:        sched.Call_ACKNOWLEDGE.Enum(),
			Acknowledge: &sched.Call_Acknowledge{
				AgentId: taskStatus.GetAgentId(),
				TaskId:  taskStatus.GetTaskId(),
				Uuid:    taskStatus.GetUuid(),
			},
		}

		h.Response.Calls = append(h.Response.Calls, call)
	}
}
