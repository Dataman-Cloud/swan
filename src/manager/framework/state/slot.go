package state

import (
	"fmt"
	"sync"

	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
)

//  TASK_STAGING = 6;  // Initial state. Framework status updates should not use.
//  TASK_STARTING = 0; // The task is being launched by the executor.
//  TASK_RUNNING = 1;
//  TASK_KILLING = 8;  // The task is being killed by the executor.
//  TASK_FINISHED = 2; // TERMINAL: The task finished successfully.
//  TASK_FAILED = 3;   // TERMINAL: The task failed to finish successfully.
//  TASK_KILLED = 4;   // TERMINAL: The task was killed by the executor.
//  TASK_ERROR = 7;    // TERMINAL: The task description contains an error.
//  TASK_LOST = 5;     // TERMINAL: The task failed but can be rescheduled.
//  TASK_DROPPED = 9;  // TERMINAL.
//  TASK_UNREACHABLE = 10;
//  TASK_GONE = 11;    // TERMINAL.
//  TASK_GONE_BY_OPERATOR = 12;
//  TASK_UNKNOWN = 13;

const (
	SLOT_STATE_PENDING_OFFER = "slot_task_pending_offer"
	SLOT_STATE_PENDING_KILL  = "slot_task_pending_killed"

	SLOT_STATE_TASK_STAGING          = "slot_task_staging"
	SLOT_STATE_TASK_STARTING         = "slot_task_starting"
	SLOT_STATE_TASK_RUNNING          = "slot_task_running"
	SLOT_STATE_TASK_KILLING          = "slot_task_killing"
	SLOT_STATE_TASK_FINISHED         = "slot_task_finished"
	SLOT_STATE_TASK_FAILED           = "slot_task_failed"
	SLOT_STATE_TASK_KILLED           = "slot_task_killed"
	SLOT_STATE_TASK_ERROR            = "slot_task_error"
	SLOT_STATE_TASK_LOST             = "slot_task_list"
	SLOT_STATE_TASK_DROPPED          = "slot_task_dropped"
	SLOT_STATE_TASK_UNREACHABLE      = "slot_task_unreachable"
	SLOT_STATE_TASK_GONE             = "slot_task_gone"
	SLOT_STATE_TASK_GONE_BY_OPERATOR = "slot_task_gone_by_operator"
	SLOT_STATE_TASK_UNKNOWN          = "slot_task_unknown"
)

type Slot struct {
	Index   int
	Id      string
	App     *App
	Version *types.Version
	State   string

	CurrentTask *Task
	TaskHistory []*Task

	OfferId       string
	AgentId       string
	Ip            string
	AgentHostName string

	resourceReservationLock sync.Mutex
}

func NewSlot(app *App, version *types.Version, index int) *Slot {
	slot := &Slot{
		Index:       index,
		App:         app,
		Version:     version,
		TaskHistory: make([]*Task, 0),
		State:       SLOT_STATE_PENDING_OFFER,
		Id:          fmt.Sprintf("%d-%s-%s-%s", index, version.AppId, version.RunAs, app.Scheduler.ClusterId), // should be app.AppId

		resourceReservationLock: sync.Mutex{},
	}

	slot.CurrentTask = NewTask(slot.App, slot.Version, slot)

	return slot
}

func (slot *Slot) TestOfferMatch(ow *OfferWrapper) bool {
	return ow.CpuRemain() > slot.Version.Cpus &&
		ow.MemRemain() > slot.Version.Mem &&
		ow.DiskRemain() > slot.Version.Disk
}

func (slot *Slot) ReserveOfferAndPrepareTaskInfo(ow *OfferWrapper) (*OfferWrapper, *mesos.TaskInfo) {
	slot.resourceReservationLock.Lock()
	defer slot.resourceReservationLock.Unlock()

	ow.CpusUsed += slot.Version.Cpus
	ow.MemUsed += slot.Version.Mem
	ow.DiskUsed += slot.Version.Disk

	return ow, slot.CurrentTask.PrepareTaskInfo(ow.Offer)
}

func (slot *Slot) Resources() []*mesos.Resource {
	var resources = []*mesos.Resource{}

	if slot.Version.Cpus > 0 {
		resources = append(resources, createScalarResource("cpus", slot.Version.Cpus))
	}

	if slot.Version.Mem > 0 {
		resources = append(resources, createScalarResource("mem", slot.Version.Cpus))
	}

	if slot.Version.Disk > 0 {
		resources = append(resources, createScalarResource("disk", slot.Version.Disk))
	}

	return resources
}

func (slot *Slot) StateIs(state string) bool {
	return slot.State == state
}

func (slot *Slot) SetState(state string) error {
	slot.State = state
	logrus.Infof("setting state for slot %s to %s", slot.Id, slot.State)

	switch slot.State {
	case SLOT_STATE_PENDING_KILL:
		slot.CurrentTask.Kill()

	case SLOT_STATE_TASK_KILLED:
	case SLOT_STATE_TASK_FINISHED:

	case SLOT_STATE_TASK_RUNNING:
	case SLOT_STATE_TASK_FAILED:
		// restart if needed
	default:
	}

	return nil
}
