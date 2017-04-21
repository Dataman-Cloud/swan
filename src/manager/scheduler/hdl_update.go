package scheduler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Dataman-Cloud/swan/src/manager/connector"
	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/state"
	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"

	"github.com/Sirupsen/logrus"
)

func UpdateHandler(s *Scheduler, ev event.Event) error {
	logrus.WithFields(logrus.Fields{"handler": "update"}).
		Debugf("logger handler report got event type: %s", ev.GetEventType())

	e, ok := ev.GetEvent().(*sched.Event)
	if !ok {
		return errUnexpectedEventType
	}

	taskStatus := e.GetUpdate().GetStatus()
	AckUpdateEvent(taskStatus)

	slotName := taskStatus.TaskId.GetValue()
	taskState := taskStatus.GetState()
	reason := taskStatus.GetReason()
	source := taskStatus.GetSource()
	message := taskStatus.GetMessage()
	healthy := taskStatus.GetHealthy()
	data := taskStatus.GetData()

	slotIndex_, appId_, userName_, clusterName_ := strings.Split(slotName, "-")[0], strings.Split(slotName, "-")[1], strings.Split(slotName, "-")[2], strings.Split(slotName, "-")[3]
	logrus.Debugf("got user name %s", userName_)
	logrus.Debugf("got cluster name %s", clusterName_)
	slotIndex, _ := strconv.ParseInt(slotIndex_, 10, 32)

	appId := fmt.Sprintf("%s-%s-%s", appId_, userName_, clusterName_)

	logrus.Debugf("got healthy report for task %s => %+v", slotName, healthy)
	logrus.Debugf("preparing set app %s slot %d to state %s", appId, slotIndex, taskState)

	app := s.AppStorage.Get(appId)
	if app == nil {
		return fmt.Errorf("app not found: %s", appId)
	}
	logrus.Debugf("found app %s", app.ID)

	slot, found := app.GetSlot(int(slotIndex))
	if !found {
		return fmt.Errorf("slot not found: %d", slotIndex)
	}
	logrus.Debugf("found slot %s", slot.ID)

	slot.SetHealthy(healthy)

	switch taskState {
	case mesos.TaskState_TASK_STAGING:
		slot.SetState(state.SLOT_STATE_TASK_STAGING)
	case mesos.TaskState_TASK_STARTING:
		slot.SetState(state.SLOT_STATE_TASK_STARTING)
	case mesos.TaskState_TASK_RUNNING:
		if !slot.StateIs(state.SLOT_STATE_TASK_RUNNING) { // set state to running only if is not previously marked as running
			slot.CurrentTask.ContainerId = parseValue(`"Id": "(?P<value>\w+)`, string(data))
			slot.CurrentTask.ContainerName = parseValue(`"Name": "(?P<value>/mesos-[\w\.-]+)`, string(data))

			slot.SetState(state.SLOT_STATE_TASK_RUNNING)
		}

	case mesos.TaskState_TASK_FINISHED:
		slot.CurrentTask.Reason = mesos.TaskStatus_Reason_name[int32(reason)]
		slot.CurrentTask.Message = message
		slot.CurrentTask.Source = mesos.TaskStatus_Source_name[int32(source)]

		slot.SetState(state.SLOT_STATE_TASK_FINISHED)

	case mesos.TaskState_TASK_FAILED:
		slot.CurrentTask.Reason = mesos.TaskStatus_Reason_name[int32(reason)]
		slot.CurrentTask.Message = message
		slot.CurrentTask.Source = mesos.TaskStatus_Source_name[int32(source)]

		slot.SetState(state.SLOT_STATE_TASK_FAILED)

	case mesos.TaskState_TASK_KILLED:
		slot.CurrentTask.Reason = mesos.TaskStatus_Reason_name[int32(reason)]
		slot.CurrentTask.Message = message
		slot.CurrentTask.Source = mesos.TaskStatus_Source_name[int32(source)]

		slot.SetState(state.SLOT_STATE_TASK_KILLED)

	case mesos.TaskState_TASK_LOST:
		slot.CurrentTask.Reason = mesos.TaskStatus_Reason_name[int32(reason)]
		slot.CurrentTask.Message = message
		slot.CurrentTask.Source = mesos.TaskStatus_Source_name[int32(source)]

		slot.SetState(state.SLOT_STATE_TASK_LOST)
	}

	return nil
}

func AckUpdateEvent(taskStatus *mesos.TaskStatus) {
	if taskStatus.GetUuid() != nil {
		call := &sched.Call{
			FrameworkId: connector.Instance().FrameworkInfo.GetId(),
			Type:        sched.Call_ACKNOWLEDGE.Enum(),
			Acknowledge: &sched.Call_Acknowledge{
				AgentId: taskStatus.GetAgentId(),
				TaskId:  taskStatus.GetTaskId(),
				Uuid:    taskStatus.GetUuid(),
			},
		}

		connector.Instance().SendCall(call)
	}
}

func parseValue(regEx, data string) string {
	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(data)

	paramsMap := make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	if value, ok := paramsMap["value"]; ok {
		return value
	}

	return ""
}

//     	TaskState_TASK_STAGING  TaskState = 6
//     	TaskState_TASK_STARTING TaskState = 0
//     	TaskState_TASK_RUNNING  TaskState = 1
//     	// NOTE: This should only be sent when the framework has
//     	// the TASK_KILLING_STATE capability.
//     	TaskState_TASK_KILLING  TaskState = 8
//     	TaskState_TASK_FINISHED TaskState = 2
//     	TaskState_TASK_FAILED   TaskState = 3
//     	TaskState_TASK_KILLED   TaskState = 4
//     	TaskState_TASK_ERROR    TaskState = 7
//     	// In Mesos 1.2, this will only be sent when the framework does NOT
//     	// opt-in to the PARTITION_AWARE capability.
//     	TaskState_TASK_LOST TaskState = 5
//     	// The task failed to launch because of a transient error. The
//     	// task's executor never started running. Unlike TASK_ERROR, the
//     	// task description is valid -- attempting to launch the task again
//     	// may be successful.
//     	TaskState_TASK_DROPPED TaskState = 9
//     	// The task was running on an agent that has lost contact with the
//     	// master, typically due to a network failure or partition. The task
//     	// may or may not still be running.
//     	TaskState_TASK_UNREACHABLE TaskState = 10
//     	// The task is no longer running. This can occur if the agent has
//     	// been terminated along with all of its tasks (e.g., the host that
//     	// was running the agent was rebooted). It might also occur if the
//     	// task was terminated due to an agent or containerizer error, or if
//     	// the task was preempted by the QoS controller in an
//     	// oversubscription scenario.
//     	TaskState_TASK_GONE TaskState = 11
//     	// The task was running on an agent that the master cannot contact;
//     	// the operator has asserted that the agent has been shutdown, but
//     	// this has not been directly confirmed by the master. If the
//     	// operator is correct, the task is not running and this is a
//     	// terminal state; if the operator is mistaken, the task may still
//     	// be running and might return to RUNNING in the future.
//     	TaskState_TASK_GONE_BY_OPERATOR TaskState = 12
//     	// The master has no knowledge of the task. This is typically
//     	// because either (a) the master never had knowledge of the task, or
//     	// (b) the master forgot about the task because it garbage collected
//     	// its metadata about the task. The task may or may not still be
//     	// running.
//     	TaskState_TASK_UNKNOWN TaskState = 13
