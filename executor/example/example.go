package example

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
	return nil
}

func (e *Executor) HandleLaunch(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Launching Task ...")

	var (
		taskInfo = ev.Launch.Task
		taskId   = taskInfo.GetTaskId()
	)

	taskbs, _ := json.MarshalIndent(taskInfo, "", "    ")
	log.Println("Mesos Executor: Task Info: ")
	fmt.Println(string(taskbs))

	driv.SendStatusUpdate(taskId, mesosproto.TaskState_TASK_STARTING.Enum(), proto.String(""))

	go func() {

		driv.SendStatusUpdate(taskId, mesosproto.TaskState_TASK_RUNNING.Enum(), proto.String(""))

		// defer to stop the executor driver
		defer driv.Stop()

		var idx int
		for {
			select {
			case <-e.stopCh:
				driv.SendStatusUpdate(taskId, mesosproto.TaskState_TASK_KILLED.Enum(), proto.String(""))
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
		taskId = ev.Kill.GetTaskId()
		policy = ev.Kill.GetKillPolicy()
	)

	log.Printf("Mesos Executor: Killing Task %s with Policy %s", taskId, policy)

	// stop the task
	driv.SendStatusUpdate(taskId, mesosproto.TaskState_TASK_KILLING.Enum(), proto.String(""))
	close(e.stopCh)
	driv.SendStatusUpdate(taskId, mesosproto.TaskState_TASK_KILLED.Enum(), proto.String(""))

	// stop the executor driver
	driv.Stop()

	return nil
}

// TODO
func (e *Executor) HandleAcknowledged(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Acknowledging ...")

	var (
		taskId = ev.Acknowledged.GetTaskId()
		uuid   = base64.StdEncoding.EncodeToString(ev.Acknowledged.GetUuid())
	)

	log.Printf("Mesos Executor: Acknowledging Task %s with uuid %s", taskId.GetValue(), uuid)
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
