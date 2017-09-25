package driver

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/Dataman-Cloud/swan/mesosproto"
	"github.com/golang/protobuf/proto"
)

var (
	MesosExecutorApiEndPoint = "/api/v1/executor"

	ErrMsgEnvNotFound = "Cannot find %s in the environment"
)

// Driver interface
//
type Driver interface {

	// EndPoint show the mesos slave executor api endpoint
	EndPoint() string

	// Starts the executor driver.
	Start() error

	// Stops the executor driver.
	Stop()

	// RegisterEventHandlers register event handlers on given executor event type
	RegisterEventHandlers(evType mesosproto.ExecEvent_Type, handleFuncs []EventHandler)

	// Sends a status update to the framework scheduler, retrying as necessary until an acknowledgement has been received
	SendStatusUpdate(taskId *mesosproto.TaskID, state *mesosproto.TaskState, message *string) error

	// Sends a message to the framework scheduler.
	SendFrameworkMessage(msg string) error
}

type EventHandler func(driv Driver, ev *mesosproto.ExecEvent) error

// DriverConfig
//
type DriverConfig struct {
	slaveID       string
	slaveEndPoint string
	slaveUPID     *UPID
	frameworkID   *mesosproto.FrameworkID
	executorID    *mesosproto.ExecutorID
	checkPoint    bool
	workDir       string
}

func (cfg *DriverConfig) String() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("slaveID=" + cfg.slaveID)
	buf.WriteString(", slaveEndPoint=" + cfg.slaveEndPoint)
	buf.WriteString(", slaveUPID=" + cfg.slaveUPID.String())
	buf.WriteString(", frameworkID=" + cfg.frameworkID.GetValue())
	buf.WriteString(", executorID=" + cfg.executorID.GetValue())
	buf.WriteString(", checkPoint=" + strconv.FormatBool(cfg.checkPoint))
	buf.WriteString(", workDir=" + cfg.workDir)
	return buf.String()
}

func NewDriverConfigFromEnv() (*DriverConfig, error) {
	var (
		cfg = new(DriverConfig)
		err error
	)

	env := os.Getenv("MESOS_SLAVE_ID")
	if env == "" {
		return nil, fmt.Errorf(ErrMsgEnvNotFound, "MESOS_SLAVE_ID")
	}
	cfg.slaveID = env

	env = os.Getenv("MESOS_AGENT_ENDPOINT")
	if env == "" {
		return nil, fmt.Errorf(ErrMsgEnvNotFound, "MESOS_AGENT_ENDPOINT")
	}
	cfg.slaveEndPoint = env

	env = os.Getenv("MESOS_FRAMEWORK_ID")
	if env == "" {
		return nil, fmt.Errorf(ErrMsgEnvNotFound, "MESOS_FRAMEWORK_ID")
	}
	cfg.frameworkID = &mesosproto.FrameworkID{
		Value: proto.String(env),
	}

	env = os.Getenv("MESOS_EXECUTOR_ID")
	if env == "" {
		return nil, fmt.Errorf(ErrMsgEnvNotFound, "MESOS_EXECUTOR_ID")
	}
	cfg.executorID = &mesosproto.ExecutorID{
		Value: proto.String(env),
	}

	env = os.Getenv("MESOS_DIRECTORY")
	if env == "" {
		return nil, fmt.Errorf(ErrMsgEnvNotFound, "MESOS_DIRECTORY")
	}
	cfg.workDir = env

	env = os.Getenv("MESOS_SLAVE_PID")
	if env == "" {
		return nil, fmt.Errorf(ErrMsgEnvNotFound, "MESOS_SLAVE_PID")
	}
	upid, err := ParseUPID(env)
	if err != nil {
		return nil, err
	}
	cfg.slaveUPID = upid

	flag, _ := strconv.ParseBool(os.Getenv("MESOS_CHECKPOINT"))
	cfg.checkPoint = flag

	return cfg, nil
}

// UPID
//
type UPID struct {
	ID   string
	Host string
	Port string
}

func ParseUPID(input string) (*UPID, error) {
	upid := new(UPID)

	if input == "" {
		return nil, fmt.Errorf("empty UPID")
	}

	splits := strings.Split(input, "@")
	if len(splits) != 2 {
		return nil, fmt.Errorf("Expect one `@' in the input")
	}
	upid.ID = splits[0]

	if _, err := net.ResolveTCPAddr("tcp4", splits[1]); err != nil {
		return nil, err
	}
	upid.Host, upid.Port, _ = net.SplitHostPort(splits[1])
	return upid, nil
}

func (u *UPID) String() string {
	return fmt.Sprintf("%s@%s:%s", u.ID, u.Host, u.Port)
}
