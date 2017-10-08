package kvm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Dataman-Cloud/swan/executor/driver"
	"github.com/Dataman-Cloud/swan/mesosproto"
	"github.com/gogo/protobuf/proto"

	log "github.com/Sirupsen/logrus"
)

type Executor struct {
	stopCh chan struct{}

	taskId  *mesosproto.TaskID // from taskinfo on Launch Event
	kvmOpts *KvmDomainOpts     // from taskinfo labels on Launch Event
}

func New() (*Executor, error) {
	return &Executor{stopCh: make(chan struct{})}, envPreCheck()
}

// TODO
// check local virt-what virt-xml-validate virsh qemu-img ...
func envPreCheck() error {
	return nil
}

func (e *Executor) HandleSubscribed(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Registered Executor on slave:", driv.EndPoint())
	return nil
}

func (e *Executor) HandlePreLaunch(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Pre Launching Task ...")

	// log the task info
	taskInfo := ev.Launch.Task
	taskbs, _ := json.MarshalIndent(taskInfo, "", "    ")
	log.Println("Mesos Executor: Task Info: ")
	fmt.Println(string(taskbs))

	// now we got the mesos task id
	e.taskId = taskInfo.GetTaskId()

	// now we got the kvm options
	// parse taskinfo labels to obtain kvm settings
	if err := e.parseKvmOptions(taskInfo); err != nil {
		log.Errorln("parse kvm options from task labels error:", err)
		return err
	}

	// fetching os iso file
	msg := e.NewMessage("IsoFetching", "fetching OS ISO file ...", "")
	e.sendMessage(driv, msg)
	time.Sleep(time.Second)
	return nil
}

func (e *Executor) HandleLaunch(driv driver.Driver, ev *mesosproto.ExecEvent) error {
	log.Println("Mesos Executor: Launching Task ...")

	// creating qemu image
	// qemu-img create -f qcow2 /var/lib/libvirt/images/centos.qcow2 10G
	msg := e.NewMessage("ImageCreating", "creating the base qemu image ...", "")
	e.sendMessage(driv, msg)
	so, se, err := RunCmd("/usr/bin/qemu-img", "create", "-f", "qcow2",
		fmt.Sprintf("/var/lib/libvirt/images/%s.qcow2", e.kvmOpts.Name),
		fmt.Sprintf("%dG", e.kvmOpts.Disk))
	cmbOutput := fmt.Sprintf("stdout=[%s], stderr=[%s]", so, se)
	fmt.Println(cmbOutput)
	if err != nil {
		e.sendUpdate(driv, mesosproto.TaskState_TASK_FAILED.Enum(), "qemu-img create error: "+err.Error())
		return err
	}

	// create the xml file
	msg = e.NewMessage("XmlCreating", "creating the xml config file ...", "")
	e.sendMessage(driv, msg)
	time.Sleep(time.Second)

	xmlbs, err := NewKvmDomain(e.kvmOpts)
	if err != nil {
		e.sendUpdate(driv, mesosproto.TaskState_TASK_FAILED.Enum(), "parse kvm options error: "+err.Error())
		return err
	}

	fileName := e.taskId.GetValue() + ".xml"
	if err := ioutil.WriteFile(fileName, xmlbs, 0644); err != nil {
		e.sendUpdate(driv, mesosproto.TaskState_TASK_FAILED.Enum(), "save kvm domain xml file error: "+err.Error())
		return err
	}

	// virsh define 4ef2002ee94a.0.demo.bbk.bj.xml
	msg = e.NewMessage("KvmDefining", "defining the kvm domain ...", "")
	e.sendMessage(driv, msg)
	so, se, err = RunCmd("/usr/bin/virsh", "define", fmt.Sprintf("--file=%s", fileName), "--validate")
	cmbOutput = fmt.Sprintf("stdout=[%s], stderr=[%s]", so, se)
	fmt.Println(cmbOutput)
	if err != nil {
		e.sendUpdate(driv, mesosproto.TaskState_TASK_FAILED.Enum(), "virsh define error: "+err.Error())
		return err
	}

	// virsh start {{.Name}}
	msg = e.NewMessage("KvmStarting", "starting the kvm domain ...", "")
	e.sendMessage(driv, msg)
	so, se, err = RunCmd("/usr/bin/virsh", "start", e.kvmOpts.Name)
	cmbOutput = fmt.Sprintf("stdout=[%s], stderr=[%s]", so, se)
	fmt.Println(cmbOutput)
	if err != nil {
		e.sendUpdate(driv, mesosproto.TaskState_TASK_FAILED.Enum(), "virsh start error: "+err.Error())
		return err
	}

	// tell the scheduler the vm vnc address
	// virsh vncdisplay bbk-demo
	so, se, err = RunCmd("/usr/bin/virsh", "vncdisplay", e.kvmOpts.Name)
	cmbOutput = fmt.Sprintf("stdout=[%s], stderr=[%s]", so, se)
	fmt.Println(cmbOutput)
	if err == nil {
		vncAddr := strings.TrimSpace(so)
		msg = e.NewMessage("KvmVncAddr", "telling vnc addr of the kvm domain ...", vncAddr)
		e.sendMessage(driv, msg)
	}

	// TODO start an goroutine to watch the vm status through virsh events
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
	// shutdown the vm & clean the base image file
	RunCmd("/usr/bin/virsh", "shutdown", e.kvmOpts.Name)
	time.Sleep(time.Second * 3)
	RunCmd("/usr/bin/virsh", "destroy", e.kvmOpts.Name)
	RunCmd("/usr/bin/virsh", "undefine", e.kvmOpts.Name)
	os.Remove(fmt.Sprintf("/var/lib/libvirt/images/%s.qcow2", e.kvmOpts.Name))
	e.sendUpdate(driv, mesosproto.TaskState_TASK_KILLED.Enum(), "")
	close(e.stopCh)

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
	var (
		msg  = ev.GetMessage() // we only care about the Message field
		data = msg.GetData()
	)

	log.Println("Mesos Executor: Message Event:")
	evbs, _ := json.MarshalIndent(msg, "", "    ")
	fmt.Println(string(evbs))

	switch msg := string(data); msg {

	case string(FMMsgShutDown):
		msg := e.NewMessage("KvmStopping", "stopping the kvm domain ...", "")
		e.sendMessage(driv, msg)

		// this require the guest vm have acpid installed
		// so, se, err := RunCmd("/usr/bin/virsh", "shutdown", e.kvmOpts.Name)

		// we use managedsave <domain> --running to perform the vm stopping
		so, se, err := RunCmd("/usr/bin/virsh", "managedsave", e.kvmOpts.Name, "--running")
		cmbOutput := fmt.Sprintf("stdout=[%s], stderr=[%s]", so, se)
		fmt.Println(cmbOutput)

		if err != nil {
			msg := e.NewMessage("KvmStopFailed", "stop the kvm domain failed: "+err.Error(), "")
			e.sendMessage(driv, msg)
		} else {
			msg := e.NewMessage("KvmStopped", "stop the kvm domain succeed", "")
			e.sendMessage(driv, msg)
		}

	case string(FMMsgStartUp):
		msg := e.NewMessage("KvmStopping", "startting the kvm domain ...", "")
		e.sendMessage(driv, msg)

		so, se, err := RunCmd("/usr/bin/virsh", "start", e.kvmOpts.Name)
		cmbOutput := fmt.Sprintf("stdout=[%s], stderr=[%s]", so, se)
		fmt.Println(cmbOutput)

		if err != nil {
			msg := e.NewMessage("KvmStartFailed", "start the kvm domain failed: "+err.Error(), "")
			e.sendMessage(driv, msg)
		} else {
			msg := e.NewMessage("KvmRunning", "start the kvm domain succeed", "")
			e.sendMessage(driv, msg)
		}

	case string(FMMsgSuspend):
		msg := e.NewMessage("KvmSuspending", "euspending the kvm domain ...", "")
		e.sendMessage(driv, msg)

		so, se, err := RunCmd("/usr/bin/virsh", "suspend", e.kvmOpts.Name)
		cmbOutput := fmt.Sprintf("stdout=[%s], stderr=[%s]", so, se)
		fmt.Println(cmbOutput)

		if err != nil {
			msg := e.NewMessage("KvmSuspendFailed", "suspend the kvm domain failed: "+err.Error(), "")
			e.sendMessage(driv, msg)
		} else {
			msg := e.NewMessage("KvmSuspended", "suspend the kvm domain succeed", "")
			e.sendMessage(driv, msg)
		}

	case string(FMMsgResume):
		msg := e.NewMessage("KvmResuming", "resuming the kvm domain ...", "")
		e.sendMessage(driv, msg)

		so, se, err := RunCmd("/usr/bin/virsh", "resume", e.kvmOpts.Name)
		cmbOutput := fmt.Sprintf("stdout=[%s], stderr=[%s]", so, se)
		fmt.Println(cmbOutput)

		if err != nil {
			msg := e.NewMessage("KvmResumeFailed", "resume the kvm domain failed: "+err.Error(), "")
			e.sendMessage(driv, msg)
		} else {
			msg := e.NewMessage("KvmRunning", "resume the kvm domain succeed", "")
			e.sendMessage(driv, msg)
		}

	default:
		log.Warnln("Mesos Executor: Unsupported kvm command message: %s, skip", msg)
	}

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

func (e *Executor) parseKvmOptions(taskInfo *mesosproto.TaskInfo) error {
	var (
		labels = taskInfo.Labels.GetLabels()
	)

	fgetlb := func(key string) string {
		for _, lb := range labels {
			if lb.GetKey() == key {
				return lb.GetValue()
			}
		}
		return ""
	}

	var (
		name        = fgetlb("SWAN_KVM_TASK_NAME")
		mem         = fgetlb("SWAN_KVM_MEMS")
		disk        = fgetlb("SWAN_KVM_DISKS")
		cpus        = fgetlb("SWAN_KVM_CPUS")
		imgUrl      = fgetlb("SWAN_KVM_IMAGE_URI")
		vncPassword = fgetlb("SWAN_KVM_VNC_PASSWORD")
		// imgTyp      = fgetlb("SWAN_KVM_IMAGE_TYPE") // TODO
	)

	cpusN, _ := strconv.Atoi(cpus)
	memN, _ := strconv.Atoi(mem)
	diskN, _ := strconv.Atoi(disk)

	// TODO valid
	// TODO support qcow2 image file

	e.kvmOpts = &KvmDomainOpts{
		Name:        name,
		Memory:      uint(memN),
		Cpus:        cpusN,
		Disk:        diskN,
		Iso:         imgUrl,
		VncPassword: vncPassword,
	}

	return nil
}

// RunCmd
func RunCmd(command string, args ...string) (string, string, error) {
	var stdoutBytes, stderrBytes bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdoutBytes
	cmd.Stderr = &stderrBytes
	err := cmd.Run()
	return stdoutBytes.String(), stderrBytes.String(), err
}

func detectExitCode(err error) (code int) {
	defer func() {
		if r := recover(); r != nil {
			code = int(math.MinInt32)
		}
	}()

	return err.(*exec.ExitError).Sys().(syscall.WaitStatus).ExitStatus()
}
