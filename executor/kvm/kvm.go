package kvm

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Dataman-Cloud/swan/executor/driver"
	"github.com/Dataman-Cloud/swan/mesosproto"
	"github.com/gogo/protobuf/proto"

	log "github.com/Sirupsen/logrus"
)

type Executor struct {
	taskId *mesosproto.TaskID
	stopCh chan struct{}
}

func New() *Executor {
	return &Executor{
		stopCh: make(chan struct{}),
	}
}

func (e *Executor) HandleSubscribed(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Registered Executor on slave:", driv.EndPoint())
	return nil
}

func (e *Executor) HandlePreLaunch(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Pre Launching Task ...")

	// log the task info
	taskInfo := ev.Launch.Task
	e.taskId = taskInfo.GetTaskId() // now we got the mesos task id

	taskbs, _ := json.MarshalIndent(taskInfo, "", "    ")
	log.Println("Mesos Executor: Task Info: ")
	fmt.Println(string(taskbs))

	// fetching os iso file
	msg := e.NewMessage("IsoFetching", "fetching OS ISO file ...")
	e.sendMessage(driv, msg)
	time.Sleep(time.Second)

	// creating qemu image
	msg = e.NewMessage("ImageCreating", "creating the base qemu image ...")
	e.sendMessage(driv, msg)
	time.Sleep(time.Second)

	// all preparing work done
	msg = e.NewMessage("Prepared", "all preparations ready")
	e.sendMessage(driv, msg)
	return nil
}

func (e *Executor) HandleLaunch(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Launching Task ...")

	// create the xml file
	msg := e.NewMessage("XmlCreating", "creating the xml config file ...")
	e.sendMessage(driv, msg)
	time.Sleep(time.Second)

	// define the xml
	msg = e.NewMessage("KvmDefining", "defining the kvm domain ...")
	e.sendMessage(driv, msg)
	time.Sleep(time.Second)

	e.sendUpdate(driv, mesosproto.TaskState_TASK_STARTING.Enum(), "")

	go func() {

		e.sendUpdate(driv, mesosproto.TaskState_TASK_RUNNING.Enum(), "")

		// defer to stop the executor driver
		defer driv.Stop()

		var idx int
		for {
			select {
			case <-e.stopCh:
				e.sendUpdate(driv, mesosproto.TaskState_TASK_KILLED.Enum(), "")
			default:
				idx++
				fmt.Println("Hello World ...", idx)
				time.Sleep(time.Second * 10)
			}
		}
	}()

	return nil
}

func (e *Executor) HandleLaunchGroup(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	return errors.New("LaunchGroup() not implemented yet")
}

func (e *Executor) HandlePreKill(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Pre Killing Task ...")
	return nil
}

// should send a terminal update (e.g., TASK_FINISHED, TASK_KILLED or TASK_FAILED)
// back to the agent once it has stopped/killed the task.
func (e *Executor) HandleKill(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Killing Task ...")

	var (
		policy = ev.Kill.GetKillPolicy()
	)

	log.Printf("Mesos Executor: Killing Task %s with Policy %s", e.taskId, policy)

	// stop the task
	e.sendUpdate(driv, mesosproto.TaskState_TASK_KILLING.Enum(), "")
	close(e.stopCh)
	e.sendUpdate(driv, mesosproto.TaskState_TASK_KILLED.Enum(), "")

	// stop the executor driver
	driv.Stop()

	return nil
}

// TODO
func (e *Executor) HandleAcknowledged(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Acknowledging ...")

	var (
		uuid = base64.StdEncoding.EncodeToString(ev.Acknowledged.GetUuid())
	)

	log.Printf("Mesos Executor: Acknowledging Task %s with uuid %s", e.taskId, uuid)
	return nil
}

// TODO It is recommended that the executor abort when it receives an error event and retry subscription.
func (e *Executor) HandleError(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Error Event:")
	evbs, _ := json.MarshalIndent(ev, "", "    ")
	fmt.Println(string(evbs))
	return nil
}

// TODO Executor should kill all its tasks, send TASK_KILLED updates and gracefully exit
func (e *Executor) HandleShutdown(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Shutdown Event:")
	evbs, _ := json.MarshalIndent(ev, "", "    ")
	fmt.Println(string(evbs))
	driv.Stop()
	return nil
}

func (e *Executor) HandleUnknown(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Unknown Event:")
	evbs, _ := json.MarshalIndent(ev, "", "    ")
	fmt.Println(string(evbs))
	return nil
}

func (e *Executor) HandleMessage(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Message Event:")
	evbs, _ := json.MarshalIndent(ev, "", "    ")
	fmt.Println(string(evbs))
	return nil
}

// send every update event with message prefixed with `SWAN_KVM_EXECUTOR_MESSAGE: `
func (e *Executor) sendUpdate(driv driver.Driver, state *mesosproto.TaskState, message string) error {
	message = "SWAN_KVM_EXECUTOR_MESSAGE: " + message
	return driv.SendStatusUpdate(e.taskId, state, proto.String(message))
}

// send every message prefixed with `SWAN_KVM_EXECUTOR_MESSAGE: `
func (e *Executor) sendMessage(driv driver.Driver, msg Message) error {
	message := "SWAN_KVM_EXECUTOR_MESSAGE: "
	bs, _ := json.Marshal(msg)
	message += string(bs)
	return driv.SendFrameworkMessage(message)
}
