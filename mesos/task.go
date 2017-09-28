package mesos

import (
	"encoding/json"
	"errors"

	"github.com/Dataman-Cloud/swan/mesosproto"
	"github.com/Dataman-Cloud/swan/types"
)

// runtime Task object
type Task struct {
	// set from Build()
	mesosproto.TaskInfo

	// for runtime mesos task updates event receive
	updates chan *mesosproto.TaskStatus

	// if set, means docker container task
	cfg *types.TaskConfig

	// if set, means kvm task
	kvmCfg *types.KvmConfig
}

// NewTask create a docker container runitme task
func NewTask(cfg *types.TaskConfig, id, name string) *Task {
	task := &Task{
		cfg:     cfg,
		updates: make(chan *mesosproto.TaskStatus, 1024),
	}

	task.Name = &name
	task.TaskId = &mesosproto.TaskID{Value: &id}

	return task
}

// NewTask create a kvm runtime task
func NewKvmTask(id, name string, cfg *types.KvmConfig) *Task {
	task := &Task{
		kvmCfg:  cfg,
		updates: make(chan *mesosproto.TaskStatus, 1024),
	}

	task.Name = &name
	task.TaskId = &mesosproto.TaskID{Value: &id}

	return task
}

func (t *Task) IsKvm() bool {
	return t.kvmCfg != nil
}

func (t *Task) ID() string {
	return t.TaskId.GetValue()
}

func (t *Task) Build() {
	if t.IsKvm() {
		t.buildKvmTask()
		return
	}
	t.buildContainerTask()
}

func (t *Task) buildKvmTask() {
	t.Resources = t.kvmCfg.BuildResources()
	t.Executor = t.kvmCfg.BuildKvmExecutor()
	t.Labels = t.kvmCfg.BuildLabels(t.ID(), t.GetName())
}

func (t *Task) buildContainerTask() {
	t.Resources = t.cfg.BuildResources()
	t.Command = t.cfg.BuildCommand()
	t.Container = t.cfg.BuildContainer(t.ID(), t.GetName())
	if t.cfg.HealthCheck != nil {
		t.HealthCheck = t.cfg.BuildHealthCheck()
	}
	t.Labels = t.cfg.BuildLabels(t.ID(), t.GetName())
}

// GetStatus method reads the task status on the updates channel
func (t *Task) GetStatus() chan *mesosproto.TaskStatus {
	return t.updates
}

// IsTaskDone check that if a task is done or not according by task status.
func IsTaskDone(status *mesosproto.TaskStatus) bool {
	state := status.GetState()

	switch state {
	case mesosproto.TaskState_TASK_RUNNING,
		mesosproto.TaskState_TASK_FINISHED,
		mesosproto.TaskState_TASK_FAILED,
		mesosproto.TaskState_TASK_KILLED,
		mesosproto.TaskState_TASK_ERROR,
		mesosproto.TaskState_TASK_LOST,
		mesosproto.TaskState_TASK_DROPPED,
		mesosproto.TaskState_TASK_UNREACHABLE,
		mesosproto.TaskState_TASK_GONE,
		mesosproto.TaskState_TASK_GONE_BY_OPERATOR,
		mesosproto.TaskState_TASK_UNKNOWN:

		return true
	}

	return false
}

func DetectTaskError(status *mesosproto.TaskStatus) error {
	var (
		state = status.GetState()
		//data  = status.GetData() // docker container inspect result
	)

	switch state {
	case mesosproto.TaskState_TASK_FAILED,
		mesosproto.TaskState_TASK_ERROR,
		mesosproto.TaskState_TASK_LOST,
		mesosproto.TaskState_TASK_DROPPED,
		mesosproto.TaskState_TASK_UNREACHABLE,
		mesosproto.TaskState_TASK_GONE,
		mesosproto.TaskState_TASK_GONE_BY_OPERATOR,
		mesosproto.TaskState_TASK_UNKNOWN:
		bs, _ := json.Marshal(map[string]interface{}{
			"message": status.GetMessage(),
			"reason":  status.GetReason().String(),
		})

		return errors.New(string(bs))
	}

	return nil
}
