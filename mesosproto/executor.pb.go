// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: executor.proto

/*
Package mesosproto is a generated protocol buffer package.

It is generated from these files:
	executor.proto
	mesos.proto
	scheduler.proto

It has these top-level messages:
	ExecEvent
	ExecCall
	FrameworkID
	OfferID
	AgentID
	TaskID
	ExecutorID
	ContainerID
	TimeInfo
	DurationInfo
	Address
	URL
	Unavailability
	MachineID
	MachineInfo
	FrameworkInfo
	HealthCheck
	KillPolicy
	CommandInfo
	ExecutorInfo
	MasterInfo
	AgentInfo
	Value
	Attribute
	Resource
	TrafficControlStatistics
	IpStatistics
	IcmpStatistics
	TcpStatistics
	UdpStatistics
	SNMPStatistics
	ResourceStatistics
	ResourceUsage
	PerfStatistics
	Request
	Offer
	InverseOffer
	TaskInfo
	TaskGroupInfo
	Task
	TaskStatus
	Filters
	Environment
	Parameter
	Parameters
	Credential
	Credentials
	RateLimit
	RateLimits
	Image
	Volume
	NetworkInfo
	CapabilityInfo
	LinuxInfo
	RLimitInfo
	TTYInfo
	ContainerInfo
	ContainerStatus
	CgroupInfo
	Labels
	Label
	Port
	Ports
	DiscoveryInfo
	WeightInfo
	VersionInfo
	Flag
	Role
	Metric
	FileInfo
	Event
	Call
*/
package mesosproto

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// Possible event types, followed by message definitions if
// applicable.
type ExecEvent_Type int32

const (
	// This must be the first enum value in this list, to
	// ensure that if 'type' is not set, the default value
	// is UNKNOWN. This enables enum values to be added
	// in a backwards-compatible way. See: MESOS-4997.
	ExecEvent_UNKNOWN      ExecEvent_Type = 0
	ExecEvent_SUBSCRIBED   ExecEvent_Type = 1
	ExecEvent_LAUNCH       ExecEvent_Type = 2
	ExecEvent_LAUNCH_GROUP ExecEvent_Type = 8
	ExecEvent_KILL         ExecEvent_Type = 3
	ExecEvent_ACKNOWLEDGED ExecEvent_Type = 4
	ExecEvent_MESSAGE      ExecEvent_Type = 5
	ExecEvent_ERROR        ExecEvent_Type = 6
	// Received when the agent asks the executor to shutdown/kill itself.
	// The executor is then required to kill all its active tasks, send
	// `TASK_KILLED` status updates and gracefully exit. The executor
	// should terminate within a `MESOS_EXECUTOR_SHUTDOWN_GRACE_PERIOD`
	// (an environment variable set by the agent upon executor startup);
	// it can be configured via `ExecutorInfo.shutdown_grace_period`. If
	// the executor fails to do so, the agent will forcefully destroy the
	// container where the executor is running. The agent would then send
	// `TASK_LOST` updates for any remaining active tasks of this executor.
	//
	// NOTE: The executor must not assume that it will always be allotted
	// the full grace period, as the agent may decide to allot a shorter
	// period and failures / forcible terminations may occur.
	//
	// TODO(alexr): Consider adding a duration field into the `Shutdown`
	// message so that the agent can communicate when a shorter period
	// has been allotted.
	ExecEvent_SHUTDOWN ExecEvent_Type = 7
)

var ExecEvent_Type_name = map[int32]string{
	0: "UNKNOWN",
	1: "SUBSCRIBED",
	2: "LAUNCH",
	8: "LAUNCH_GROUP",
	3: "KILL",
	4: "ACKNOWLEDGED",
	5: "MESSAGE",
	6: "ERROR",
	7: "SHUTDOWN",
}
var ExecEvent_Type_value = map[string]int32{
	"UNKNOWN":      0,
	"SUBSCRIBED":   1,
	"LAUNCH":       2,
	"LAUNCH_GROUP": 8,
	"KILL":         3,
	"ACKNOWLEDGED": 4,
	"MESSAGE":      5,
	"ERROR":        6,
	"SHUTDOWN":     7,
}

func (x ExecEvent_Type) Enum() *ExecEvent_Type {
	p := new(ExecEvent_Type)
	*p = x
	return p
}
func (x ExecEvent_Type) String() string {
	return proto.EnumName(ExecEvent_Type_name, int32(x))
}
func (x *ExecEvent_Type) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(ExecEvent_Type_value, data, "ExecEvent_Type")
	if err != nil {
		return err
	}
	*x = ExecEvent_Type(value)
	return nil
}
func (ExecEvent_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptorExecutor, []int{0, 0} }

// Possible call types, followed by message definitions if
// applicable.
type ExecCall_Type int32

const (
	// See comments above on `Event::Type` for more details on this enum value.
	ExecCall_UNKNOWN   ExecCall_Type = 0
	ExecCall_SUBSCRIBE ExecCall_Type = 1
	ExecCall_UPDATE    ExecCall_Type = 2
	ExecCall_MESSAGE   ExecCall_Type = 3
)

var ExecCall_Type_name = map[int32]string{
	0: "UNKNOWN",
	1: "SUBSCRIBE",
	2: "UPDATE",
	3: "MESSAGE",
}
var ExecCall_Type_value = map[string]int32{
	"UNKNOWN":   0,
	"SUBSCRIBE": 1,
	"UPDATE":    2,
	"MESSAGE":   3,
}

func (x ExecCall_Type) Enum() *ExecCall_Type {
	p := new(ExecCall_Type)
	*p = x
	return p
}
func (x ExecCall_Type) String() string {
	return proto.EnumName(ExecCall_Type_name, int32(x))
}
func (x *ExecCall_Type) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(ExecCall_Type_value, data, "ExecCall_Type")
	if err != nil {
		return err
	}
	*x = ExecCall_Type(value)
	return nil
}
func (ExecCall_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptorExecutor, []int{1, 0} }

// *
// Executor event API.
//
// An event is described using the standard protocol buffer "union"
// trick, see https://developers.google.com/protocol-buffers/docs/techniques#union.
type ExecEvent struct {
	// Type of the event, indicates which optional field below should be
	// present if that type has a nested message definition.
	// Enum fields should be optional, see: MESOS-4997.
	Type             *ExecEvent_Type         `protobuf:"varint,1,opt,name=type,enum=mesos.executor.ExecEvent_Type" json:"type,omitempty"`
	Subscribed       *ExecEvent_Subscribed   `protobuf:"bytes,2,opt,name=subscribed" json:"subscribed,omitempty"`
	Acknowledged     *ExecEvent_Acknowledged `protobuf:"bytes,3,opt,name=acknowledged" json:"acknowledged,omitempty"`
	Launch           *ExecEvent_Launch       `protobuf:"bytes,4,opt,name=launch" json:"launch,omitempty"`
	LaunchGroup      *ExecEvent_LaunchGroup  `protobuf:"bytes,8,opt,name=launch_group" json:"launch_group,omitempty"`
	Kill             *ExecEvent_Kill         `protobuf:"bytes,5,opt,name=kill" json:"kill,omitempty"`
	Message          *ExecEvent_Message      `protobuf:"bytes,6,opt,name=message" json:"message,omitempty"`
	Error            *ExecEvent_Error        `protobuf:"bytes,7,opt,name=error" json:"error,omitempty"`
	XXX_unrecognized []byte                  `json:"-"`
}

func (m *ExecEvent) Reset()                    { *m = ExecEvent{} }
func (m *ExecEvent) String() string            { return proto.CompactTextString(m) }
func (*ExecEvent) ProtoMessage()               {}
func (*ExecEvent) Descriptor() ([]byte, []int) { return fileDescriptorExecutor, []int{0} }

func (m *ExecEvent) GetType() ExecEvent_Type {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return ExecEvent_UNKNOWN
}

func (m *ExecEvent) GetSubscribed() *ExecEvent_Subscribed {
	if m != nil {
		return m.Subscribed
	}
	return nil
}

func (m *ExecEvent) GetAcknowledged() *ExecEvent_Acknowledged {
	if m != nil {
		return m.Acknowledged
	}
	return nil
}

func (m *ExecEvent) GetLaunch() *ExecEvent_Launch {
	if m != nil {
		return m.Launch
	}
	return nil
}

func (m *ExecEvent) GetLaunchGroup() *ExecEvent_LaunchGroup {
	if m != nil {
		return m.LaunchGroup
	}
	return nil
}

func (m *ExecEvent) GetKill() *ExecEvent_Kill {
	if m != nil {
		return m.Kill
	}
	return nil
}

func (m *ExecEvent) GetMessage() *ExecEvent_Message {
	if m != nil {
		return m.Message
	}
	return nil
}

func (m *ExecEvent) GetError() *ExecEvent_Error {
	if m != nil {
		return m.Error
	}
	return nil
}

// First event received when the executor subscribes.
// The 'id' field in the 'framework_info' will be set.
type ExecEvent_Subscribed struct {
	ExecutorInfo  *ExecutorInfo  `protobuf:"bytes,1,req,name=executor_info" json:"executor_info,omitempty"`
	FrameworkInfo *FrameworkInfo `protobuf:"bytes,2,req,name=framework_info" json:"framework_info,omitempty"`
	AgentInfo     *AgentInfo     `protobuf:"bytes,3,req,name=agent_info" json:"agent_info,omitempty"`
	// Uniquely identifies the container of an executor run.
	ContainerId      *ContainerID `protobuf:"bytes,4,opt,name=container_id" json:"container_id,omitempty"`
	XXX_unrecognized []byte       `json:"-"`
}

func (m *ExecEvent_Subscribed) Reset()                    { *m = ExecEvent_Subscribed{} }
func (m *ExecEvent_Subscribed) String() string            { return proto.CompactTextString(m) }
func (*ExecEvent_Subscribed) ProtoMessage()               {}
func (*ExecEvent_Subscribed) Descriptor() ([]byte, []int) { return fileDescriptorExecutor, []int{0, 0} }

func (m *ExecEvent_Subscribed) GetExecutorInfo() *ExecutorInfo {
	if m != nil {
		return m.ExecutorInfo
	}
	return nil
}

func (m *ExecEvent_Subscribed) GetFrameworkInfo() *FrameworkInfo {
	if m != nil {
		return m.FrameworkInfo
	}
	return nil
}

func (m *ExecEvent_Subscribed) GetAgentInfo() *AgentInfo {
	if m != nil {
		return m.AgentInfo
	}
	return nil
}

func (m *ExecEvent_Subscribed) GetContainerId() *ContainerID {
	if m != nil {
		return m.ContainerId
	}
	return nil
}

// Received when the framework attempts to launch a task. Once
// the task is successfully launched, the executor must respond with
// a TASK_RUNNING update (See TaskState in v1/mesos.proto).
type ExecEvent_Launch struct {
	Task             *TaskInfo `protobuf:"bytes,1,req,name=task" json:"task,omitempty"`
	XXX_unrecognized []byte    `json:"-"`
}

func (m *ExecEvent_Launch) Reset()                    { *m = ExecEvent_Launch{} }
func (m *ExecEvent_Launch) String() string            { return proto.CompactTextString(m) }
func (*ExecEvent_Launch) ProtoMessage()               {}
func (*ExecEvent_Launch) Descriptor() ([]byte, []int) { return fileDescriptorExecutor, []int{0, 1} }

func (m *ExecEvent_Launch) GetTask() *TaskInfo {
	if m != nil {
		return m.Task
	}
	return nil
}

// Received when the framework attempts to launch a group of tasks atomically.
// Similar to `Launch` above the executor must send TASK_RUNNING updates for
// tasks that are successfully launched.
type ExecEvent_LaunchGroup struct {
	TaskGroup        *TaskGroupInfo `protobuf:"bytes,1,req,name=task_group" json:"task_group,omitempty"`
	XXX_unrecognized []byte         `json:"-"`
}

func (m *ExecEvent_LaunchGroup) Reset()                    { *m = ExecEvent_LaunchGroup{} }
func (m *ExecEvent_LaunchGroup) String() string            { return proto.CompactTextString(m) }
func (*ExecEvent_LaunchGroup) ProtoMessage()               {}
func (*ExecEvent_LaunchGroup) Descriptor() ([]byte, []int) { return fileDescriptorExecutor, []int{0, 2} }

func (m *ExecEvent_LaunchGroup) GetTaskGroup() *TaskGroupInfo {
	if m != nil {
		return m.TaskGroup
	}
	return nil
}

// Received when the scheduler wants to kill a specific task. Once
// the task is terminated, the executor should send a TASK_KILLED
// (or TASK_FAILED) update. The terminal update is necessary so
// Mesos can release the resources associated with the task.
type ExecEvent_Kill struct {
	TaskId *TaskID `protobuf:"bytes,1,req,name=task_id" json:"task_id,omitempty"`
	// If set, overrides any previously specified kill policy for this task.
	// This includes 'TaskInfo.kill_policy' and 'Executor.kill.kill_policy'.
	// Can be used to forcefully kill a task which is already being killed.
	KillPolicy       *KillPolicy `protobuf:"bytes,2,opt,name=kill_policy" json:"kill_policy,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *ExecEvent_Kill) Reset()                    { *m = ExecEvent_Kill{} }
func (m *ExecEvent_Kill) String() string            { return proto.CompactTextString(m) }
func (*ExecEvent_Kill) ProtoMessage()               {}
func (*ExecEvent_Kill) Descriptor() ([]byte, []int) { return fileDescriptorExecutor, []int{0, 3} }

func (m *ExecEvent_Kill) GetTaskId() *TaskID {
	if m != nil {
		return m.TaskId
	}
	return nil
}

func (m *ExecEvent_Kill) GetKillPolicy() *KillPolicy {
	if m != nil {
		return m.KillPolicy
	}
	return nil
}

// Received when the agent acknowledges the receipt of status
// update. Schedulers are responsible for explicitly acknowledging
// the receipt of status updates that have 'update.status().uuid()'
// field set. Unacknowledged updates can be retried by the executor.
// They should also be sent by the executor whenever it
// re-subscribes.
type ExecEvent_Acknowledged struct {
	TaskId           *TaskID `protobuf:"bytes,1,req,name=task_id" json:"task_id,omitempty"`
	Uuid             []byte  `protobuf:"bytes,2,req,name=uuid" json:"uuid,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *ExecEvent_Acknowledged) Reset()         { *m = ExecEvent_Acknowledged{} }
func (m *ExecEvent_Acknowledged) String() string { return proto.CompactTextString(m) }
func (*ExecEvent_Acknowledged) ProtoMessage()    {}
func (*ExecEvent_Acknowledged) Descriptor() ([]byte, []int) {
	return fileDescriptorExecutor, []int{0, 4}
}

func (m *ExecEvent_Acknowledged) GetTaskId() *TaskID {
	if m != nil {
		return m.TaskId
	}
	return nil
}

func (m *ExecEvent_Acknowledged) GetUuid() []byte {
	if m != nil {
		return m.Uuid
	}
	return nil
}

// Received when a custom message generated by the scheduler is
// forwarded by the agent. Note that this message is not
// interpreted by Mesos and is only forwarded (without reliability
// guarantees) to the executor. It is up to the scheduler to retry
// if the message is dropped for any reason.
type ExecEvent_Message struct {
	Data             []byte `protobuf:"bytes,1,req,name=data" json:"data,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *ExecEvent_Message) Reset()                    { *m = ExecEvent_Message{} }
func (m *ExecEvent_Message) String() string            { return proto.CompactTextString(m) }
func (*ExecEvent_Message) ProtoMessage()               {}
func (*ExecEvent_Message) Descriptor() ([]byte, []int) { return fileDescriptorExecutor, []int{0, 5} }

func (m *ExecEvent_Message) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

// Received in case the executor sends invalid calls (e.g.,
// required values not set).
// TODO(arojas): Remove this once the old executor driver is no
// longer supported. With HTTP API all errors will be signaled via
// HTTP response codes.
type ExecEvent_Error struct {
	Message          *string `protobuf:"bytes,1,req,name=message" json:"message,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *ExecEvent_Error) Reset()                    { *m = ExecEvent_Error{} }
func (m *ExecEvent_Error) String() string            { return proto.CompactTextString(m) }
func (*ExecEvent_Error) ProtoMessage()               {}
func (*ExecEvent_Error) Descriptor() ([]byte, []int) { return fileDescriptorExecutor, []int{0, 6} }

func (m *ExecEvent_Error) GetMessage() string {
	if m != nil && m.Message != nil {
		return *m.Message
	}
	return ""
}

// *
// Executor call API.
//
// Like Event, a Call is described using the standard protocol buffer
// "union" trick (see above).
type ExecCall struct {
	// Identifies the executor which generated this call.
	ExecutorId  *ExecutorID  `protobuf:"bytes,1,req,name=executor_id" json:"executor_id,omitempty"`
	FrameworkId *FrameworkID `protobuf:"bytes,2,req,name=framework_id" json:"framework_id,omitempty"`
	// Type of the call, indicates which optional field below should be
	// present if that type has a nested message definition.
	// In case type is SUBSCRIBED, no message needs to be set.
	// See comments on `Event::Type` above on the reasoning behind this
	// field being optional.
	Type             *ExecCall_Type      `protobuf:"varint,3,opt,name=type,enum=mesos.executor.ExecCall_Type" json:"type,omitempty"`
	Subscribe        *ExecCall_Subscribe `protobuf:"bytes,4,opt,name=subscribe" json:"subscribe,omitempty"`
	Update           *ExecCall_Update    `protobuf:"bytes,5,opt,name=update" json:"update,omitempty"`
	Message          *ExecCall_Message   `protobuf:"bytes,6,opt,name=message" json:"message,omitempty"`
	XXX_unrecognized []byte              `json:"-"`
}

func (m *ExecCall) Reset()                    { *m = ExecCall{} }
func (m *ExecCall) String() string            { return proto.CompactTextString(m) }
func (*ExecCall) ProtoMessage()               {}
func (*ExecCall) Descriptor() ([]byte, []int) { return fileDescriptorExecutor, []int{1} }

func (m *ExecCall) GetExecutorId() *ExecutorID {
	if m != nil {
		return m.ExecutorId
	}
	return nil
}

func (m *ExecCall) GetFrameworkId() *FrameworkID {
	if m != nil {
		return m.FrameworkId
	}
	return nil
}

func (m *ExecCall) GetType() ExecCall_Type {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return ExecCall_UNKNOWN
}

func (m *ExecCall) GetSubscribe() *ExecCall_Subscribe {
	if m != nil {
		return m.Subscribe
	}
	return nil
}

func (m *ExecCall) GetUpdate() *ExecCall_Update {
	if m != nil {
		return m.Update
	}
	return nil
}

func (m *ExecCall) GetMessage() *ExecCall_Message {
	if m != nil {
		return m.Message
	}
	return nil
}

// Request to subscribe with the agent. If subscribing after a disconnection,
// it must include a list of all the tasks and updates which haven't been
// acknowledged by the scheduler.
type ExecCall_Subscribe struct {
	UnacknowledgedTasks   []*TaskInfo        `protobuf:"bytes,1,rep,name=unacknowledged_tasks" json:"unacknowledged_tasks,omitempty"`
	UnacknowledgedUpdates []*ExecCall_Update `protobuf:"bytes,2,rep,name=unacknowledged_updates" json:"unacknowledged_updates,omitempty"`
	XXX_unrecognized      []byte             `json:"-"`
}

func (m *ExecCall_Subscribe) Reset()                    { *m = ExecCall_Subscribe{} }
func (m *ExecCall_Subscribe) String() string            { return proto.CompactTextString(m) }
func (*ExecCall_Subscribe) ProtoMessage()               {}
func (*ExecCall_Subscribe) Descriptor() ([]byte, []int) { return fileDescriptorExecutor, []int{1, 0} }

func (m *ExecCall_Subscribe) GetUnacknowledgedTasks() []*TaskInfo {
	if m != nil {
		return m.UnacknowledgedTasks
	}
	return nil
}

func (m *ExecCall_Subscribe) GetUnacknowledgedUpdates() []*ExecCall_Update {
	if m != nil {
		return m.UnacknowledgedUpdates
	}
	return nil
}

// Notifies the scheduler that a task has transitioned from one
// state to another. Status updates should be used by executors
// to reliably communicate the status of the tasks that they
// manage. It is crucial that a terminal update (see TaskState
// in v1/mesos.proto) is sent to the scheduler as soon as the task
// terminates, in order for Mesos to release the resources allocated
// to the task. It is the responsibility of the scheduler to
// explicitly acknowledge the receipt of a status update. See
// 'Acknowledged' in the 'Events' section above for the semantics.
type ExecCall_Update struct {
	Status           *TaskStatus `protobuf:"bytes,1,req,name=status" json:"status,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *ExecCall_Update) Reset()                    { *m = ExecCall_Update{} }
func (m *ExecCall_Update) String() string            { return proto.CompactTextString(m) }
func (*ExecCall_Update) ProtoMessage()               {}
func (*ExecCall_Update) Descriptor() ([]byte, []int) { return fileDescriptorExecutor, []int{1, 1} }

func (m *ExecCall_Update) GetStatus() *TaskStatus {
	if m != nil {
		return m.Status
	}
	return nil
}

// Sends arbitrary binary data to the scheduler. Note that Mesos
// neither interprets this data nor makes any guarantees about the
// delivery of this message to the scheduler.
// See 'Message' in the 'Events' section.
type ExecCall_Message struct {
	Data             []byte `protobuf:"bytes,2,req,name=data" json:"data,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *ExecCall_Message) Reset()                    { *m = ExecCall_Message{} }
func (m *ExecCall_Message) String() string            { return proto.CompactTextString(m) }
func (*ExecCall_Message) ProtoMessage()               {}
func (*ExecCall_Message) Descriptor() ([]byte, []int) { return fileDescriptorExecutor, []int{1, 2} }

func (m *ExecCall_Message) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*ExecEvent)(nil), "mesos.executor.ExecEvent")
	proto.RegisterType((*ExecEvent_Subscribed)(nil), "mesos.executor.ExecEvent.Subscribed")
	proto.RegisterType((*ExecEvent_Launch)(nil), "mesos.executor.ExecEvent.Launch")
	proto.RegisterType((*ExecEvent_LaunchGroup)(nil), "mesos.executor.ExecEvent.LaunchGroup")
	proto.RegisterType((*ExecEvent_Kill)(nil), "mesos.executor.ExecEvent.Kill")
	proto.RegisterType((*ExecEvent_Acknowledged)(nil), "mesos.executor.ExecEvent.Acknowledged")
	proto.RegisterType((*ExecEvent_Message)(nil), "mesos.executor.ExecEvent.Message")
	proto.RegisterType((*ExecEvent_Error)(nil), "mesos.executor.ExecEvent.Error")
	proto.RegisterType((*ExecCall)(nil), "mesos.executor.ExecCall")
	proto.RegisterType((*ExecCall_Subscribe)(nil), "mesos.executor.ExecCall.Subscribe")
	proto.RegisterType((*ExecCall_Update)(nil), "mesos.executor.ExecCall.Update")
	proto.RegisterType((*ExecCall_Message)(nil), "mesos.executor.ExecCall.Message")
	proto.RegisterEnum("mesos.executor.ExecEvent_Type", ExecEvent_Type_name, ExecEvent_Type_value)
	proto.RegisterEnum("mesos.executor.ExecCall_Type", ExecCall_Type_name, ExecCall_Type_value)
}

func init() { proto.RegisterFile("executor.proto", fileDescriptorExecutor) }

var fileDescriptorExecutor = []byte{
	// 741 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x95, 0x7f, 0x4f, 0xda, 0x5c,
	0x14, 0xc7, 0x9f, 0x42, 0x29, 0x70, 0xa8, 0x58, 0xef, 0x63, 0x9e, 0xa7, 0x69, 0xa2, 0x22, 0x71,
	0x8e, 0x4c, 0xc7, 0x36, 0x92, 0x65, 0x4b, 0x34, 0x59, 0x90, 0x76, 0x48, 0x44, 0x34, 0xfc, 0xc8,
	0x92, 0xfd, 0x43, 0x2a, 0xbd, 0xb2, 0x86, 0xda, 0x92, 0xfe, 0x98, 0x9a, 0xfd, 0xb9, 0xb7, 0xb0,
	0x97, 0xb0, 0xd7, 0xb0, 0xd7, 0xb7, 0xdc, 0xd3, 0x0b, 0x22, 0xa3, 0xf3, 0x3f, 0xe5, 0x7c, 0x3e,
	0xf7, 0x5c, 0xee, 0xf9, 0xde, 0x0b, 0x14, 0xe9, 0x1d, 0x1d, 0x45, 0xa1, 0xe7, 0x57, 0xa7, 0xbe,
	0x17, 0x7a, 0xa4, 0x78, 0x43, 0x03, 0x2f, 0xa8, 0xce, 0x3e, 0xd5, 0x0a, 0xf1, 0xff, 0x58, 0x2c,
	0xff, 0xc8, 0x41, 0xde, 0xb8, 0xa3, 0x23, 0xe3, 0x2b, 0x75, 0x43, 0x72, 0x08, 0x62, 0x78, 0x3f,
	0xa5, 0xaa, 0x50, 0x12, 0x2a, 0xc5, 0xda, 0x76, 0xf5, 0xb1, 0x59, 0x9d, 0x83, 0xd5, 0xfe, 0xfd,
	0x94, 0x92, 0xf7, 0x00, 0x41, 0x74, 0x15, 0x8c, 0x7c, 0xfb, 0x8a, 0x5a, 0x6a, 0xaa, 0x24, 0x54,
	0x0a, 0xb5, 0xbd, 0x64, 0xa7, 0x37, 0x67, 0xc9, 0x31, 0xc8, 0xe6, 0x68, 0xe2, 0x7a, 0xb7, 0x0e,
	0xb5, 0xc6, 0xd4, 0x52, 0xd3, 0xe8, 0xee, 0x27, 0xbb, 0xf5, 0x05, 0x9a, 0xbc, 0x06, 0xc9, 0x31,
	0x23, 0x77, 0xf4, 0x45, 0x15, 0xd1, 0x2b, 0x25, 0x7b, 0x6d, 0xe4, 0xc8, 0x11, 0xc8, 0xb1, 0x31,
	0x1c, 0xfb, 0x5e, 0x34, 0x55, 0x73, 0xe8, 0x3d, 0x7b, 0xca, 0x6b, 0x32, 0x98, 0x1d, 0xca, 0xc4,
	0x76, 0x1c, 0x35, 0x83, 0xd2, 0x5f, 0x0e, 0xe5, 0xcc, 0x76, 0x1c, 0x52, 0x83, 0xec, 0x0d, 0x0d,
	0x02, 0x73, 0x4c, 0x55, 0x09, 0x85, 0xdd, 0x64, 0xe1, 0x3c, 0x06, 0x49, 0x15, 0x32, 0xd4, 0xf7,
	0x3d, 0x5f, 0xcd, 0xa2, 0xb1, 0x93, 0x6c, 0x18, 0x0c, 0xd3, 0x7e, 0x09, 0x00, 0x0b, 0xa7, 0xf9,
	0x02, 0xd6, 0x66, 0xe8, 0xd0, 0x76, 0xaf, 0x3d, 0x55, 0x28, 0xa5, 0x2a, 0x85, 0xda, 0xbf, 0x7c,
	0x19, 0x83, 0xd7, 0x5a, 0xee, 0xb5, 0x47, 0x0e, 0xa1, 0x78, 0xed, 0x9b, 0x37, 0xf4, 0xd6, 0xf3,
	0x27, 0x31, 0x9c, 0x42, 0x78, 0x93, 0xc3, 0x1f, 0x67, 0x45, 0xa4, 0xf7, 0x00, 0xcc, 0x31, 0x75,
	0xc3, 0x98, 0x4c, 0x23, 0xa9, 0x70, 0xb2, 0xce, 0x0a, 0x48, 0x55, 0x40, 0x1e, 0x79, 0x6e, 0x68,
	0xda, 0x2e, 0xf5, 0x87, 0xb6, 0xc5, 0xa7, 0x42, 0x38, 0xd7, 0x98, 0x95, 0x5a, 0xba, 0xf6, 0x1c,
	0x24, 0x3e, 0x91, 0x2d, 0x10, 0x43, 0x33, 0x98, 0xf0, 0xad, 0xae, 0x73, 0xb6, 0x6f, 0x06, 0xd8,
	0x58, 0x7b, 0x07, 0x85, 0xc5, 0x11, 0x54, 0x00, 0x18, 0xcd, 0xa7, 0x27, 0x3c, 0xda, 0x31, 0x73,
	0x90, 0x42, 0xb1, 0x03, 0x22, 0x8e, 0x61, 0x1b, 0xb2, 0x68, 0xd8, 0x16, 0xc7, 0xd7, 0x16, 0x5b,
	0xe8, 0x64, 0x1f, 0x0a, 0x6c, 0xa8, 0xc3, 0xa9, 0xe7, 0xd8, 0xa3, 0x7b, 0x1e, 0xde, 0x0d, 0xce,
	0xb0, 0x15, 0x2e, 0xb1, 0xa0, 0x1d, 0x83, 0xfc, 0x28, 0x7b, 0x4f, 0xad, 0x2b, 0x83, 0x18, 0x45,
	0xb6, 0x85, 0xa7, 0x2a, 0x6b, 0xff, 0x43, 0x76, 0x36, 0x63, 0x19, 0x44, 0xcb, 0x0c, 0x4d, 0xb4,
	0x64, 0x4d, 0x85, 0x0c, 0x8e, 0x92, 0xac, 0x3f, 0xc4, 0x85, 0x55, 0xf2, 0xe5, 0xef, 0x02, 0x88,
	0x78, 0xbb, 0x0a, 0x90, 0x1d, 0x74, 0xce, 0x3a, 0x17, 0x9f, 0x3a, 0xca, 0x3f, 0xa4, 0x08, 0xd0,
	0x1b, 0x9c, 0xf4, 0x1a, 0xdd, 0xd6, 0x89, 0xa1, 0x2b, 0x02, 0x01, 0x90, 0xda, 0xf5, 0x41, 0xa7,
	0x71, 0xaa, 0xa4, 0x88, 0x02, 0x72, 0xfc, 0xf7, 0xb0, 0xd9, 0xbd, 0x18, 0x5c, 0x2a, 0x39, 0x92,
	0x03, 0xf1, 0xac, 0xd5, 0x6e, 0x2b, 0x69, 0x56, 0xab, 0x37, 0xd8, 0x22, 0x6d, 0x43, 0x6f, 0x1a,
	0xba, 0x22, 0xb2, 0x65, 0xcf, 0x8d, 0x5e, 0xaf, 0xde, 0x34, 0x94, 0x0c, 0xc9, 0x43, 0xc6, 0xe8,
	0x76, 0x2f, 0xba, 0x8a, 0x44, 0x64, 0xc8, 0xf5, 0x4e, 0x07, 0x7d, 0x9d, 0xf5, 0xcb, 0x96, 0x7f,
	0x8a, 0x90, 0x63, 0xb9, 0x69, 0x98, 0x8e, 0xc3, 0xce, 0xea, 0x21, 0x5f, 0xb3, 0xef, 0xbd, 0xb1,
	0x9c, 0x2e, 0x9d, 0xe5, 0x60, 0x21, 0x5b, 0x16, 0x4f, 0x16, 0xf9, 0x23, 0x59, 0x3a, 0x39, 0xe0,
	0xef, 0x4c, 0x1a, 0xdf, 0x99, 0xad, 0x55, 0x79, 0x67, 0x9d, 0xe3, 0x67, 0xe6, 0x2d, 0xe4, 0xe7,
	0xcf, 0x0c, 0xcf, 0x56, 0x39, 0xd1, 0x98, 0x5f, 0x0b, 0xf2, 0x0a, 0xa4, 0x68, 0x6a, 0x99, 0x21,
	0xe5, 0x17, 0x77, 0x27, 0xd1, 0x19, 0x20, 0x46, 0xde, 0x2c, 0xdf, 0xdc, 0x52, 0xa2, 0xc1, 0x87,
	0xaa, 0x7d, 0x83, 0xfc, 0x43, 0xc3, 0x97, 0xb0, 0x19, 0xb9, 0x8b, 0xcf, 0xda, 0x90, 0x25, 0x25,
	0x50, 0x85, 0x52, 0x7a, 0x45, 0xc4, 0xc9, 0x07, 0xf8, 0x6f, 0x09, 0x8f, 0xb7, 0x1b, 0xa8, 0x29,
	0x14, 0x9e, 0xda, 0xaf, 0x76, 0x00, 0x12, 0xdf, 0xf9, 0x2e, 0x48, 0x41, 0x68, 0x86, 0x51, 0xb0,
	0x34, 0x1b, 0xd6, 0xab, 0x87, 0x85, 0x55, 0x49, 0xc4, 0x88, 0x96, 0x8f, 0x56, 0xc5, 0x6d, 0x0d,
	0xf2, 0xf3, 0xb8, 0xc5, 0x69, 0x1b, 0x5c, 0xea, 0xf5, 0xbe, 0xa1, 0xa4, 0x16, 0xf3, 0x93, 0x3e,
	0x91, 0x3f, 0x03, 0x76, 0xc2, 0xdf, 0x92, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x5c, 0x33, 0xa5,
	0x70, 0x79, 0x06, 0x00, 0x00,
}
