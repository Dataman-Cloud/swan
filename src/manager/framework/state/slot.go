package state

import (
	"fmt"
	"sync"

	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/types"
)

const (
	SLOT_STATE_PENDING_OFFER     = "slot_task_pending_offer"
	SLOT_STATE_TASK_DISPATCHED   = "slot_task_dispatched"
	SLOT_STATE_TASK_RUNNING      = "slot_task_running"
	SLOT_STATE_TASK_FAILED       = "slot_task_need_reschedue"
	SLOT_STATE_TASK_RESCHEDULING = "slot_task_rescheduling"
	SLOT_STATE_TASK_STOPPING     = "slot_task_stopping"
	SLOT_STATE_TASK_STOPPED      = "slot_task_stopped"
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
		Id:          fmt.Sprintf("%d-%s-%s-%s", index, version.AppId, version.RunAs, app.ClusterId), // should be app.AppId

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

func (slot *Slot) SetState(state string) {
	slot.State = state

	switch slot.State {
	case SLOT_STATE_TASK_DISPATCHED:
		fmt.Println(slot.State)
	case SLOT_STATE_TASK_RUNNING:
		fmt.Println(slot.State)
	case SLOT_STATE_TASK_FAILED:
		fmt.Println(slot.State)
		// restart if needed
	default:
	}
	// persist to db
}
