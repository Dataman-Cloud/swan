// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: scheduler.proto

package mesosproto

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Possible event types, followed by message definitions if
// applicable.
type Event_Type int32

const (
	// This must be the first enum value in this list, to
	// ensure that if 'type' is not set, the default value
	// is UNKNOWN. This enables enum values to be added
	// in a backwards-compatible way. See: MESOS-4997.
	Event_UNKNOWN    Event_Type = 0
	Event_SUBSCRIBED Event_Type = 1
	Event_RESCIND    Event_Type = 3
	Event_UPDATE     Event_Type = 4
	Event_MESSAGE    Event_Type = 5
	Event_FAILURE    Event_Type = 6
	Event_OFFERS     Event_Type = 2
	Event_ERROR      Event_Type = 7
	// Periodic message sent by the Mesos master according to
	// 'Subscribed.heartbeat_interval_seconds'. If the scheduler does
	// not receive any events (including heartbeats) for an extended
	// period of time (e.g., 5 x heartbeat_interval_seconds), there is
	// likely a network partition. In such a case the scheduler should
	// close the existing subscription connection and resubscribe
	// using a backoff strategy.
	Event_HEARTBEAT Event_Type = 8
)

var Event_Type_name = map[int32]string{
	0: "UNKNOWN",
	1: "SUBSCRIBED",
	3: "RESCIND",
	4: "UPDATE",
	5: "MESSAGE",
	6: "FAILURE",
	2: "OFFERS",
	7: "ERROR",
	8: "HEARTBEAT",
}
var Event_Type_value = map[string]int32{
	"UNKNOWN":    0,
	"SUBSCRIBED": 1,
	"RESCIND":    3,
	"UPDATE":     4,
	"MESSAGE":    5,
	"FAILURE":    6,
	"OFFERS":     2,
	"ERROR":      7,
	"HEARTBEAT":  8,
}

func (x Event_Type) Enum() *Event_Type {
	p := new(Event_Type)
	*p = x
	return p
}
func (x Event_Type) String() string {
	return proto.EnumName(Event_Type_name, int32(x))
}
func (x *Event_Type) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(Event_Type_value, data, "Event_Type")
	if err != nil {
		return err
	}
	*x = Event_Type(value)
	return nil
}
func (Event_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{0, 0} }

// Possible call types, followed by message definitions if
// applicable.
type Call_Type int32

const (
	// See comments above on `Event::Type` for more details on this enum value.
	Call_UNKNOWN     Call_Type = 0
	Call_SUBSCRIBE   Call_Type = 1
	Call_TEARDOWN    Call_Type = 2
	Call_ACCEPT      Call_Type = 3
	Call_DECLINE     Call_Type = 4
	Call_REVIVE      Call_Type = 5
	Call_KILL        Call_Type = 6
	Call_SHUTDOWN    Call_Type = 7
	Call_ACKNOWLEDGE Call_Type = 8
	Call_RECONCILE   Call_Type = 9
	Call_MESSAGE     Call_Type = 10
	Call_REQUEST     Call_Type = 11
	Call_SUPPRESS    Call_Type = 12
)

var Call_Type_name = map[int32]string{
	0:  "UNKNOWN",
	1:  "SUBSCRIBE",
	2:  "TEARDOWN",
	3:  "ACCEPT",
	4:  "DECLINE",
	5:  "REVIVE",
	6:  "KILL",
	7:  "SHUTDOWN",
	8:  "ACKNOWLEDGE",
	9:  "RECONCILE",
	10: "MESSAGE",
	11: "REQUEST",
	12: "SUPPRESS",
}
var Call_Type_value = map[string]int32{
	"UNKNOWN":     0,
	"SUBSCRIBE":   1,
	"TEARDOWN":    2,
	"ACCEPT":      3,
	"DECLINE":     4,
	"REVIVE":      5,
	"KILL":        6,
	"SHUTDOWN":    7,
	"ACKNOWLEDGE": 8,
	"RECONCILE":   9,
	"MESSAGE":     10,
	"REQUEST":     11,
	"SUPPRESS":    12,
}

func (x Call_Type) Enum() *Call_Type {
	p := new(Call_Type)
	*p = x
	return p
}
func (x Call_Type) String() string {
	return proto.EnumName(Call_Type_name, int32(x))
}
func (x *Call_Type) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(Call_Type_value, data, "Call_Type")
	if err != nil {
		return err
	}
	*x = Call_Type(value)
	return nil
}
func (Call_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{1, 0} }

// *
// Scheduler event API.
//
// An event is described using the standard protocol buffer "union"
// trick, see:
// https://developers.google.com/protocol-buffers/docs/techniques#union.
type Event struct {
	// Type of the event, indicates which optional field below should be
	// present if that type has a nested message definition.
	// Enum fields should be optional, see: MESOS-4997.
	Type             *Event_Type       `protobuf:"varint,1,opt,name=type,enum=mesos.scheduler.Event_Type" json:"type,omitempty"`
	Subscribed       *Event_Subscribed `protobuf:"bytes,2,opt,name=subscribed" json:"subscribed,omitempty"`
	Offers           *Event_Offers     `protobuf:"bytes,3,opt,name=offers" json:"offers,omitempty"`
	Rescind          *Event_Rescind    `protobuf:"bytes,4,opt,name=rescind" json:"rescind,omitempty"`
	Update           *Event_Update     `protobuf:"bytes,5,opt,name=update" json:"update,omitempty"`
	Message          *Event_Message    `protobuf:"bytes,6,opt,name=message" json:"message,omitempty"`
	Failure          *Event_Failure    `protobuf:"bytes,7,opt,name=failure" json:"failure,omitempty"`
	Error            *Event_Error      `protobuf:"bytes,8,opt,name=error" json:"error,omitempty"`
	XXX_unrecognized []byte            `json:"-"`
}

func (m *Event) Reset()                    { *m = Event{} }
func (m *Event) String() string            { return proto.CompactTextString(m) }
func (*Event) ProtoMessage()               {}
func (*Event) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{0} }

func (m *Event) GetType() Event_Type {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return Event_UNKNOWN
}

func (m *Event) GetSubscribed() *Event_Subscribed {
	if m != nil {
		return m.Subscribed
	}
	return nil
}

func (m *Event) GetOffers() *Event_Offers {
	if m != nil {
		return m.Offers
	}
	return nil
}

func (m *Event) GetRescind() *Event_Rescind {
	if m != nil {
		return m.Rescind
	}
	return nil
}

func (m *Event) GetUpdate() *Event_Update {
	if m != nil {
		return m.Update
	}
	return nil
}

func (m *Event) GetMessage() *Event_Message {
	if m != nil {
		return m.Message
	}
	return nil
}

func (m *Event) GetFailure() *Event_Failure {
	if m != nil {
		return m.Failure
	}
	return nil
}

func (m *Event) GetError() *Event_Error {
	if m != nil {
		return m.Error
	}
	return nil
}

// First event received when the scheduler subscribes.
type Event_Subscribed struct {
	FrameworkId *FrameworkID `protobuf:"bytes,1,req,name=framework_id" json:"framework_id,omitempty"`
	// This value will be set if the master is sending heartbeats. See
	// the comment above on 'HEARTBEAT' for more details.
	HeartbeatIntervalSeconds *float64 `protobuf:"fixed64,2,opt,name=heartbeat_interval_seconds" json:"heartbeat_interval_seconds,omitempty"`
	XXX_unrecognized         []byte   `json:"-"`
}

func (m *Event_Subscribed) Reset()                    { *m = Event_Subscribed{} }
func (m *Event_Subscribed) String() string            { return proto.CompactTextString(m) }
func (*Event_Subscribed) ProtoMessage()               {}
func (*Event_Subscribed) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{0, 0} }

func (m *Event_Subscribed) GetFrameworkId() *FrameworkID {
	if m != nil {
		return m.FrameworkId
	}
	return nil
}

func (m *Event_Subscribed) GetHeartbeatIntervalSeconds() float64 {
	if m != nil && m.HeartbeatIntervalSeconds != nil {
		return *m.HeartbeatIntervalSeconds
	}
	return 0
}

// Received whenever there are new resources that are offered to the
// scheduler or resources requested back from the scheduler. Each
// offer corresponds to a set of resources on an agent. Until the
// scheduler accepts or declines an offer the resources are
// considered allocated to the scheduler. Accepting or Declining an
// inverse offer informs the allocator of the scheduler's ability to
// release the resources without violating an SLA.
type Event_Offers struct {
	Offers           []*Offer        `protobuf:"bytes,1,rep,name=offers" json:"offers,omitempty"`
	InverseOffers    []*InverseOffer `protobuf:"bytes,2,rep,name=inverse_offers" json:"inverse_offers,omitempty"`
	XXX_unrecognized []byte          `json:"-"`
}

func (m *Event_Offers) Reset()                    { *m = Event_Offers{} }
func (m *Event_Offers) String() string            { return proto.CompactTextString(m) }
func (*Event_Offers) ProtoMessage()               {}
func (*Event_Offers) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{0, 1} }

func (m *Event_Offers) GetOffers() []*Offer {
	if m != nil {
		return m.Offers
	}
	return nil
}

func (m *Event_Offers) GetInverseOffers() []*InverseOffer {
	if m != nil {
		return m.InverseOffers
	}
	return nil
}

// Received when a particular offer is no longer valid (e.g., the
// agent corresponding to the offer has been removed) and hence
// needs to be rescinded. Any future calls ('Accept' / 'Decline') made
// by the scheduler regarding this offer will be invalid.
type Event_Rescind struct {
	OfferId          *OfferID `protobuf:"bytes,1,req,name=offer_id" json:"offer_id,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *Event_Rescind) Reset()                    { *m = Event_Rescind{} }
func (m *Event_Rescind) String() string            { return proto.CompactTextString(m) }
func (*Event_Rescind) ProtoMessage()               {}
func (*Event_Rescind) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{0, 2} }

func (m *Event_Rescind) GetOfferId() *OfferID {
	if m != nil {
		return m.OfferId
	}
	return nil
}

// Received whenever there is a status update that is generated by
// the executor or agent or master. Status updates should be used by
// executors to reliably communicate the status of the tasks that
// they manage. It is crucial that a terminal update (see TaskState
// in v1/mesos.proto) is sent by the executor as soon as the task
// terminates, in order for Mesos to release the resources allocated
// to the task. It is also the responsibility of the scheduler to
// explicitly acknowledge the receipt of a status update. See
// 'Acknowledge' in the 'Call' section below for the semantics.
type Event_Update struct {
	Status           *TaskStatus `protobuf:"bytes,1,req,name=status" json:"status,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *Event_Update) Reset()                    { *m = Event_Update{} }
func (m *Event_Update) String() string            { return proto.CompactTextString(m) }
func (*Event_Update) ProtoMessage()               {}
func (*Event_Update) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{0, 3} }

func (m *Event_Update) GetStatus() *TaskStatus {
	if m != nil {
		return m.Status
	}
	return nil
}

// Received when a custom message generated by the executor is
// forwarded by the master. Note that this message is not
// interpreted by Mesos and is only forwarded (without reliability
// guarantees) to the scheduler. It is up to the executor to retry
// if the message is dropped for any reason.
type Event_Message struct {
	AgentId          *AgentID    `protobuf:"bytes,1,req,name=agent_id" json:"agent_id,omitempty"`
	ExecutorId       *ExecutorID `protobuf:"bytes,2,req,name=executor_id" json:"executor_id,omitempty"`
	Data             []byte      `protobuf:"bytes,3,req,name=data" json:"data,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *Event_Message) Reset()                    { *m = Event_Message{} }
func (m *Event_Message) String() string            { return proto.CompactTextString(m) }
func (*Event_Message) ProtoMessage()               {}
func (*Event_Message) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{0, 4} }

func (m *Event_Message) GetAgentId() *AgentID {
	if m != nil {
		return m.AgentId
	}
	return nil
}

func (m *Event_Message) GetExecutorId() *ExecutorID {
	if m != nil {
		return m.ExecutorId
	}
	return nil
}

func (m *Event_Message) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

// Received when an agent is removed from the cluster (e.g., failed
// health checks) or when an executor is terminated. Note that, this
// event coincides with receipt of terminal UPDATE events for any
// active tasks belonging to the agent or executor and receipt of
// 'Rescind' events for any outstanding offers belonging to the
// agent. Note that there is no guaranteed order between the
// 'Failure', 'Update' and 'Rescind' events when an agent or executor
// is removed.
// TODO(vinod): Consider splitting the lost agent and terminated
// executor into separate events and ensure it's reliably generated.
type Event_Failure struct {
	AgentId *AgentID `protobuf:"bytes,1,opt,name=agent_id" json:"agent_id,omitempty"`
	// If this was just a failure of an executor on an agent then
	// 'executor_id' will be set and possibly 'status' (if we were
	// able to determine the exit status).
	ExecutorId       *ExecutorID `protobuf:"bytes,2,opt,name=executor_id" json:"executor_id,omitempty"`
	Status           *int32      `protobuf:"varint,3,opt,name=status" json:"status,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *Event_Failure) Reset()                    { *m = Event_Failure{} }
func (m *Event_Failure) String() string            { return proto.CompactTextString(m) }
func (*Event_Failure) ProtoMessage()               {}
func (*Event_Failure) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{0, 5} }

func (m *Event_Failure) GetAgentId() *AgentID {
	if m != nil {
		return m.AgentId
	}
	return nil
}

func (m *Event_Failure) GetExecutorId() *ExecutorID {
	if m != nil {
		return m.ExecutorId
	}
	return nil
}

func (m *Event_Failure) GetStatus() int32 {
	if m != nil && m.Status != nil {
		return *m.Status
	}
	return 0
}

// Received when there is an unrecoverable error in the scheduler (e.g.,
// scheduler failed over, rate limiting, authorization errors etc.). The
// scheduler should abort on receiving this event.
type Event_Error struct {
	Message          *string `protobuf:"bytes,1,req,name=message" json:"message,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Event_Error) Reset()                    { *m = Event_Error{} }
func (m *Event_Error) String() string            { return proto.CompactTextString(m) }
func (*Event_Error) ProtoMessage()               {}
func (*Event_Error) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{0, 6} }

func (m *Event_Error) GetMessage() string {
	if m != nil && m.Message != nil {
		return *m.Message
	}
	return ""
}

// *
// Scheduler call API.
//
// Like Event, a Call is described using the standard protocol buffer
// "union" trick (see above).
type Call struct {
	// Identifies who generated this call. Master assigns a framework id
	// when a new scheduler subscribes for the first time. Once assigned,
	// the scheduler must set the 'framework_id' here and within its
	// FrameworkInfo (in any further 'Subscribe' calls). This allows the
	// master to identify a scheduler correctly across disconnections,
	// failovers, etc.
	FrameworkId *FrameworkID `protobuf:"bytes,1,opt,name=framework_id" json:"framework_id,omitempty"`
	// Type of the call, indicates which optional field below should be
	// present if that type has a nested message definition.
	// See comments on `Event::Type` above on the reasoning behind this field being optional.
	Type             *Call_Type        `protobuf:"varint,2,opt,name=type,enum=mesos.scheduler.Call_Type" json:"type,omitempty"`
	Subscribe        *Call_Subscribe   `protobuf:"bytes,3,opt,name=subscribe" json:"subscribe,omitempty"`
	Accept           *Call_Accept      `protobuf:"bytes,4,opt,name=accept" json:"accept,omitempty"`
	Decline          *Call_Decline     `protobuf:"bytes,5,opt,name=decline" json:"decline,omitempty"`
	Kill             *Call_Kill        `protobuf:"bytes,6,opt,name=kill" json:"kill,omitempty"`
	Shutdown         *Call_Shutdown    `protobuf:"bytes,7,opt,name=shutdown" json:"shutdown,omitempty"`
	Acknowledge      *Call_Acknowledge `protobuf:"bytes,8,opt,name=acknowledge" json:"acknowledge,omitempty"`
	Reconcile        *Call_Reconcile   `protobuf:"bytes,9,opt,name=reconcile" json:"reconcile,omitempty"`
	Message          *Call_Message     `protobuf:"bytes,10,opt,name=message" json:"message,omitempty"`
	Request          *Call_Request     `protobuf:"bytes,11,opt,name=request" json:"request,omitempty"`
	XXX_unrecognized []byte            `json:"-"`
}

func (m *Call) Reset()                    { *m = Call{} }
func (m *Call) String() string            { return proto.CompactTextString(m) }
func (*Call) ProtoMessage()               {}
func (*Call) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{1} }

func (m *Call) GetFrameworkId() *FrameworkID {
	if m != nil {
		return m.FrameworkId
	}
	return nil
}

func (m *Call) GetType() Call_Type {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return Call_UNKNOWN
}

func (m *Call) GetSubscribe() *Call_Subscribe {
	if m != nil {
		return m.Subscribe
	}
	return nil
}

func (m *Call) GetAccept() *Call_Accept {
	if m != nil {
		return m.Accept
	}
	return nil
}

func (m *Call) GetDecline() *Call_Decline {
	if m != nil {
		return m.Decline
	}
	return nil
}

func (m *Call) GetKill() *Call_Kill {
	if m != nil {
		return m.Kill
	}
	return nil
}

func (m *Call) GetShutdown() *Call_Shutdown {
	if m != nil {
		return m.Shutdown
	}
	return nil
}

func (m *Call) GetAcknowledge() *Call_Acknowledge {
	if m != nil {
		return m.Acknowledge
	}
	return nil
}

func (m *Call) GetReconcile() *Call_Reconcile {
	if m != nil {
		return m.Reconcile
	}
	return nil
}

func (m *Call) GetMessage() *Call_Message {
	if m != nil {
		return m.Message
	}
	return nil
}

func (m *Call) GetRequest() *Call_Request {
	if m != nil {
		return m.Request
	}
	return nil
}

// Subscribes the scheduler with the master to receive events. A
// scheduler must send other calls only after it has received the
// SUBCRIBED event.
type Call_Subscribe struct {
	// See the comments below on 'framework_id' on the semantics for
	// 'framework_info.id'.
	FrameworkInfo    *FrameworkInfo `protobuf:"bytes,1,req,name=framework_info" json:"framework_info,omitempty"`
	XXX_unrecognized []byte         `json:"-"`
}

func (m *Call_Subscribe) Reset()                    { *m = Call_Subscribe{} }
func (m *Call_Subscribe) String() string            { return proto.CompactTextString(m) }
func (*Call_Subscribe) ProtoMessage()               {}
func (*Call_Subscribe) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{1, 0} }

func (m *Call_Subscribe) GetFrameworkInfo() *FrameworkInfo {
	if m != nil {
		return m.FrameworkInfo
	}
	return nil
}

// Accepts an offer, performing the specified operations
// in a sequential manner.
//
// E.g. Launch a task with a newly reserved persistent volume:
//
//   Accept {
//     offer_ids: [ ... ]
//     operations: [
//       { type: RESERVE,
//         reserve: { resources: [ disk(role):2 ] } }
//       { type: CREATE,
//         create: { volumes: [ disk(role):1+persistence ] } }
//       { type: LAUNCH,
//         launch: { task_infos ... disk(role):1;disk(role):1+persistence } }
//     ]
//   }
//
// Note that any of the offer’s resources not used in the 'Accept'
// call (e.g., to launch a task) are considered unused and might be
// reoffered to other frameworks. In other words, the same OfferID
// cannot be used in more than one 'Accept' call.
type Call_Accept struct {
	OfferIds         []*OfferID         `protobuf:"bytes,1,rep,name=offer_ids" json:"offer_ids,omitempty"`
	Operations       []*Offer_Operation `protobuf:"bytes,2,rep,name=operations" json:"operations,omitempty"`
	Filters          *Filters           `protobuf:"bytes,3,opt,name=filters" json:"filters,omitempty"`
	XXX_unrecognized []byte             `json:"-"`
}

func (m *Call_Accept) Reset()                    { *m = Call_Accept{} }
func (m *Call_Accept) String() string            { return proto.CompactTextString(m) }
func (*Call_Accept) ProtoMessage()               {}
func (*Call_Accept) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{1, 1} }

func (m *Call_Accept) GetOfferIds() []*OfferID {
	if m != nil {
		return m.OfferIds
	}
	return nil
}

func (m *Call_Accept) GetOperations() []*Offer_Operation {
	if m != nil {
		return m.Operations
	}
	return nil
}

func (m *Call_Accept) GetFilters() *Filters {
	if m != nil {
		return m.Filters
	}
	return nil
}

// Declines an offer, signaling the master to potentially reoffer
// the resources to a different framework. Note that this is same
// as sending an Accept call with no operations. See comments on
// top of 'Accept' for semantics.
type Call_Decline struct {
	OfferIds         []*OfferID `protobuf:"bytes,1,rep,name=offer_ids" json:"offer_ids,omitempty"`
	Filters          *Filters   `protobuf:"bytes,2,opt,name=filters" json:"filters,omitempty"`
	XXX_unrecognized []byte     `json:"-"`
}

func (m *Call_Decline) Reset()                    { *m = Call_Decline{} }
func (m *Call_Decline) String() string            { return proto.CompactTextString(m) }
func (*Call_Decline) ProtoMessage()               {}
func (*Call_Decline) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{1, 2} }

func (m *Call_Decline) GetOfferIds() []*OfferID {
	if m != nil {
		return m.OfferIds
	}
	return nil
}

func (m *Call_Decline) GetFilters() *Filters {
	if m != nil {
		return m.Filters
	}
	return nil
}

// Kills a specific task. If the scheduler has a custom executor,
// the kill is forwarded to the executor and it is up to the
// executor to kill the task and send a TASK_KILLED (or TASK_FAILED)
// update. Note that Mesos releases the resources for a task once it
// receives a terminal update (See TaskState in v1/mesos.proto) for
// it. If the task is unknown to the master, a TASK_LOST update is
// generated.
type Call_Kill struct {
	TaskId  *TaskID  `protobuf:"bytes,1,req,name=task_id" json:"task_id,omitempty"`
	AgentId *AgentID `protobuf:"bytes,2,opt,name=agent_id" json:"agent_id,omitempty"`
	// If set, overrides any previously specified kill policy for this task.
	// This includes 'TaskInfo.kill_policy' and 'Executor.kill.kill_policy'.
	// Can be used to forcefully kill a task which is already being killed.
	KillPolicy       *KillPolicy `protobuf:"bytes,3,opt,name=kill_policy" json:"kill_policy,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *Call_Kill) Reset()                    { *m = Call_Kill{} }
func (m *Call_Kill) String() string            { return proto.CompactTextString(m) }
func (*Call_Kill) ProtoMessage()               {}
func (*Call_Kill) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{1, 3} }

func (m *Call_Kill) GetTaskId() *TaskID {
	if m != nil {
		return m.TaskId
	}
	return nil
}

func (m *Call_Kill) GetAgentId() *AgentID {
	if m != nil {
		return m.AgentId
	}
	return nil
}

func (m *Call_Kill) GetKillPolicy() *KillPolicy {
	if m != nil {
		return m.KillPolicy
	}
	return nil
}

// Shuts down a custom executor. When the executor gets a shutdown
// event, it is expected to kill all its tasks (and send TASK_KILLED
// updates) and terminate. If the executor doesn’t terminate within
// a certain timeout (configurable via
// '--executor_shutdown_grace_period' agent flag), the agent will
// forcefully destroy the container (executor and its tasks) and
// transition its active tasks to TASK_LOST.
type Call_Shutdown struct {
	ExecutorId       *ExecutorID `protobuf:"bytes,1,req,name=executor_id" json:"executor_id,omitempty"`
	AgentId          *AgentID    `protobuf:"bytes,2,req,name=agent_id" json:"agent_id,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *Call_Shutdown) Reset()                    { *m = Call_Shutdown{} }
func (m *Call_Shutdown) String() string            { return proto.CompactTextString(m) }
func (*Call_Shutdown) ProtoMessage()               {}
func (*Call_Shutdown) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{1, 4} }

func (m *Call_Shutdown) GetExecutorId() *ExecutorID {
	if m != nil {
		return m.ExecutorId
	}
	return nil
}

func (m *Call_Shutdown) GetAgentId() *AgentID {
	if m != nil {
		return m.AgentId
	}
	return nil
}

// Acknowledges the receipt of status update. Schedulers are
// responsible for explicitly acknowledging the receipt of status
// updates that have 'Update.status().uuid()' field set. Such status
// updates are retried by the agent until they are acknowledged by
// the scheduler.
type Call_Acknowledge struct {
	AgentId          *AgentID `protobuf:"bytes,1,req,name=agent_id" json:"agent_id,omitempty"`
	TaskId           *TaskID  `protobuf:"bytes,2,req,name=task_id" json:"task_id,omitempty"`
	Uuid             []byte   `protobuf:"bytes,3,req,name=uuid" json:"uuid,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *Call_Acknowledge) Reset()                    { *m = Call_Acknowledge{} }
func (m *Call_Acknowledge) String() string            { return proto.CompactTextString(m) }
func (*Call_Acknowledge) ProtoMessage()               {}
func (*Call_Acknowledge) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{1, 5} }

func (m *Call_Acknowledge) GetAgentId() *AgentID {
	if m != nil {
		return m.AgentId
	}
	return nil
}

func (m *Call_Acknowledge) GetTaskId() *TaskID {
	if m != nil {
		return m.TaskId
	}
	return nil
}

func (m *Call_Acknowledge) GetUuid() []byte {
	if m != nil {
		return m.Uuid
	}
	return nil
}

// Allows the scheduler to query the status for non-terminal tasks.
// This causes the master to send back the latest task status for
// each task in 'tasks', if possible. Tasks that are no longer known
// will result in a TASK_LOST update. If 'statuses' is empty, then
// the master will send the latest status for each task currently
// known.
type Call_Reconcile struct {
	Tasks            []*Call_Reconcile_Task `protobuf:"bytes,1,rep,name=tasks" json:"tasks,omitempty"`
	XXX_unrecognized []byte                 `json:"-"`
}

func (m *Call_Reconcile) Reset()                    { *m = Call_Reconcile{} }
func (m *Call_Reconcile) String() string            { return proto.CompactTextString(m) }
func (*Call_Reconcile) ProtoMessage()               {}
func (*Call_Reconcile) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{1, 6} }

func (m *Call_Reconcile) GetTasks() []*Call_Reconcile_Task {
	if m != nil {
		return m.Tasks
	}
	return nil
}

// TODO(vinod): Support arbitrary queries than just state of tasks.
type Call_Reconcile_Task struct {
	TaskId           *TaskID  `protobuf:"bytes,1,req,name=task_id" json:"task_id,omitempty"`
	AgentId          *AgentID `protobuf:"bytes,2,opt,name=agent_id" json:"agent_id,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *Call_Reconcile_Task) Reset()         { *m = Call_Reconcile_Task{} }
func (m *Call_Reconcile_Task) String() string { return proto.CompactTextString(m) }
func (*Call_Reconcile_Task) ProtoMessage()    {}
func (*Call_Reconcile_Task) Descriptor() ([]byte, []int) {
	return fileDescriptorScheduler, []int{1, 6, 0}
}

func (m *Call_Reconcile_Task) GetTaskId() *TaskID {
	if m != nil {
		return m.TaskId
	}
	return nil
}

func (m *Call_Reconcile_Task) GetAgentId() *AgentID {
	if m != nil {
		return m.AgentId
	}
	return nil
}

// Sends arbitrary binary data to the executor. Note that Mesos
// neither interprets this data nor makes any guarantees about the
// delivery of this message to the executor.
type Call_Message struct {
	AgentId          *AgentID    `protobuf:"bytes,1,req,name=agent_id" json:"agent_id,omitempty"`
	ExecutorId       *ExecutorID `protobuf:"bytes,2,req,name=executor_id" json:"executor_id,omitempty"`
	Data             []byte      `protobuf:"bytes,3,req,name=data" json:"data,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *Call_Message) Reset()                    { *m = Call_Message{} }
func (m *Call_Message) String() string            { return proto.CompactTextString(m) }
func (*Call_Message) ProtoMessage()               {}
func (*Call_Message) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{1, 7} }

func (m *Call_Message) GetAgentId() *AgentID {
	if m != nil {
		return m.AgentId
	}
	return nil
}

func (m *Call_Message) GetExecutorId() *ExecutorID {
	if m != nil {
		return m.ExecutorId
	}
	return nil
}

func (m *Call_Message) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

// Requests a specific set of resources from Mesos's allocator. If
// the allocator has support for this, corresponding offers will be
// sent asynchronously via the OFFERS event(s).
//
// NOTE: The built-in hierarchical allocator doesn't have support
// for this call and hence simply ignores it.
type Call_Request struct {
	Requests         []*Request `protobuf:"bytes,1,rep,name=requests" json:"requests,omitempty"`
	XXX_unrecognized []byte     `json:"-"`
}

func (m *Call_Request) Reset()                    { *m = Call_Request{} }
func (m *Call_Request) String() string            { return proto.CompactTextString(m) }
func (*Call_Request) ProtoMessage()               {}
func (*Call_Request) Descriptor() ([]byte, []int) { return fileDescriptorScheduler, []int{1, 8} }

func (m *Call_Request) GetRequests() []*Request {
	if m != nil {
		return m.Requests
	}
	return nil
}

func init() {
	proto.RegisterType((*Event)(nil), "mesos.scheduler.Event")
	proto.RegisterType((*Event_Subscribed)(nil), "mesos.scheduler.Event.Subscribed")
	proto.RegisterType((*Event_Offers)(nil), "mesos.scheduler.Event.Offers")
	proto.RegisterType((*Event_Rescind)(nil), "mesos.scheduler.Event.Rescind")
	proto.RegisterType((*Event_Update)(nil), "mesos.scheduler.Event.Update")
	proto.RegisterType((*Event_Message)(nil), "mesos.scheduler.Event.Message")
	proto.RegisterType((*Event_Failure)(nil), "mesos.scheduler.Event.Failure")
	proto.RegisterType((*Event_Error)(nil), "mesos.scheduler.Event.Error")
	proto.RegisterType((*Call)(nil), "mesos.scheduler.Call")
	proto.RegisterType((*Call_Subscribe)(nil), "mesos.scheduler.Call.Subscribe")
	proto.RegisterType((*Call_Accept)(nil), "mesos.scheduler.Call.Accept")
	proto.RegisterType((*Call_Decline)(nil), "mesos.scheduler.Call.Decline")
	proto.RegisterType((*Call_Kill)(nil), "mesos.scheduler.Call.Kill")
	proto.RegisterType((*Call_Shutdown)(nil), "mesos.scheduler.Call.Shutdown")
	proto.RegisterType((*Call_Acknowledge)(nil), "mesos.scheduler.Call.Acknowledge")
	proto.RegisterType((*Call_Reconcile)(nil), "mesos.scheduler.Call.Reconcile")
	proto.RegisterType((*Call_Reconcile_Task)(nil), "mesos.scheduler.Call.Reconcile.Task")
	proto.RegisterType((*Call_Message)(nil), "mesos.scheduler.Call.Message")
	proto.RegisterType((*Call_Request)(nil), "mesos.scheduler.Call.Request")
	proto.RegisterEnum("mesos.scheduler.Event_Type", Event_Type_name, Event_Type_value)
	proto.RegisterEnum("mesos.scheduler.Call_Type", Call_Type_name, Call_Type_value)
}

func init() { proto.RegisterFile("scheduler.proto", fileDescriptorScheduler) }

var fileDescriptorScheduler = []byte{
	// 1026 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x96, 0xdd, 0x6e, 0xe3, 0x44,
	0x14, 0xc7, 0xb1, 0x9b, 0xd8, 0xc9, 0x71, 0x9a, 0x9a, 0x01, 0x21, 0xcb, 0x94, 0xdd, 0xb4, 0x42,
	0x28, 0x50, 0x08, 0xa8, 0x08, 0x24, 0x2e, 0xd3, 0x78, 0xb2, 0xb5, 0x9a, 0x26, 0xc5, 0x76, 0x16,
	0x69, 0x25, 0x54, 0xb9, 0xf6, 0xa4, 0xb5, 0xea, 0xda, 0xc1, 0x1f, 0xed, 0x56, 0xe2, 0x09, 0xb8,
	0xe6, 0x55, 0x78, 0x31, 0x9e, 0x00, 0xcd, 0x78, 0x9c, 0xba, 0x6d, 0x9c, 0x5d, 0x09, 0x69, 0xaf,
	0xaa, 0xfa, 0xfc, 0xce, 0x39, 0xf3, 0xf1, 0x3f, 0xff, 0x09, 0xec, 0xa4, 0xde, 0x15, 0xf1, 0xf3,
	0x90, 0x24, 0x83, 0x65, 0x12, 0x67, 0x31, 0xda, 0xb9, 0x21, 0x69, 0x9c, 0x0e, 0x56, 0x9f, 0x75,
	0xa5, 0xf8, 0xc0, 0xa2, 0xfb, 0xff, 0xca, 0xd0, 0xc4, 0xb7, 0x24, 0xca, 0xd0, 0xd7, 0xd0, 0xc8,
	0xee, 0x97, 0x44, 0x13, 0x7a, 0x42, 0xbf, 0x7b, 0xf8, 0xf9, 0xe0, 0x49, 0xda, 0x80, 0x51, 0x03,
	0xe7, 0x7e, 0x49, 0xd0, 0x4f, 0x00, 0x69, 0x7e, 0x91, 0x7a, 0x49, 0x70, 0x41, 0x7c, 0x4d, 0xec,
	0x09, 0x7d, 0xe5, 0x70, 0xaf, 0x26, 0xc1, 0x5e, 0x81, 0xe8, 0x3b, 0x90, 0xe2, 0xc5, 0x82, 0x24,
	0xa9, 0xb6, 0xc5, 0x52, 0xbe, 0xa8, 0x49, 0x99, 0x31, 0x08, 0x7d, 0x0f, 0x72, 0x42, 0x52, 0x2f,
	0x88, 0x7c, 0xad, 0xc1, 0xf8, 0x17, 0x35, 0xbc, 0x55, 0x50, 0xb4, 0x7e, 0xbe, 0xf4, 0xdd, 0x8c,
	0x68, 0xcd, 0x8d, 0xf5, 0xe7, 0x0c, 0xa2, 0xf5, 0x6f, 0x48, 0x9a, 0xba, 0x97, 0x44, 0x93, 0x36,
	0xd6, 0x3f, 0x2d, 0x28, 0x9a, 0xb0, 0x70, 0x83, 0x30, 0x4f, 0x88, 0x26, 0x6f, 0x4c, 0x18, 0x17,
	0x14, 0x3a, 0x80, 0x26, 0x49, 0x92, 0x38, 0xd1, 0x5a, 0x0c, 0xdf, 0xad, 0xc1, 0x31, 0x65, 0xf4,
	0x37, 0x00, 0x95, 0xb3, 0xea, 0x43, 0x67, 0x91, 0xb8, 0x37, 0xe4, 0x2e, 0x4e, 0xae, 0xcf, 0x03,
	0x5f, 0x13, 0x7a, 0x62, 0x5f, 0x39, 0x44, 0xbc, 0xc2, 0xb8, 0x0c, 0x99, 0x06, 0xda, 0x07, 0xfd,
	0x8a, 0xb8, 0x49, 0x76, 0x41, 0xdc, 0xec, 0x3c, 0x88, 0x32, 0x92, 0xdc, 0xba, 0xe1, 0x79, 0x4a,
	0xbc, 0x38, 0xf2, 0x53, 0x76, 0x39, 0x82, 0x6e, 0x83, 0xc4, 0x0f, 0x75, 0x77, 0x75, 0x07, 0x42,
	0x6f, 0xab, 0xaf, 0x1c, 0x76, 0x78, 0x45, 0x16, 0x46, 0x07, 0xd0, 0x0d, 0xa2, 0x5b, 0x92, 0xa4,
	0xe4, 0x9c, 0x53, 0x22, 0xa3, 0x3e, 0xe1, 0x94, 0x59, 0x04, 0x19, 0xac, 0x1f, 0x80, 0x5c, 0x9e,
	0x7c, 0x0f, 0x5a, 0x8c, 0x7f, 0x58, 0x69, 0xb7, 0x5a, 0xd7, 0x34, 0xf4, 0x03, 0x90, 0xf8, 0xb1,
	0xef, 0x81, 0x94, 0x66, 0x6e, 0x96, 0xa7, 0x9c, 0xfc, 0x98, 0x93, 0x8e, 0x9b, 0x5e, 0xdb, 0x2c,
	0xa0, 0xbb, 0x20, 0x97, 0x67, 0xde, 0x83, 0x96, 0x7b, 0x49, 0xa2, 0xec, 0x79, 0xe5, 0x21, 0xfd,
	0x6c, 0x1a, 0xe8, 0x2b, 0x50, 0xc8, 0x5b, 0xe2, 0xe5, 0x59, 0xcc, 0xda, 0x8b, 0x8f, 0x8a, 0x62,
	0x1e, 0x31, 0x0d, 0xd4, 0x81, 0x86, 0xef, 0x66, 0xae, 0xb6, 0xd5, 0x13, 0xfb, 0x1d, 0xdd, 0x03,
	0xb9, 0xbc, 0xa5, 0xc7, 0x2d, 0x84, 0xf7, 0x69, 0x21, 0xac, 0x6f, 0xd1, 0x5d, 0x6d, 0x8d, 0x0a,
	0xbc, 0xa9, 0x6b, 0xd0, 0x64, 0x77, 0x8b, 0x76, 0x1e, 0xa4, 0x46, 0x37, 0xd1, 0xde, 0xff, 0x13,
	0x1a, 0x6c, 0x92, 0x14, 0x90, 0xe7, 0xd3, 0x93, 0xe9, 0xec, 0xb7, 0xa9, 0xfa, 0x11, 0xea, 0x02,
	0xd8, 0xf3, 0x23, 0x7b, 0x64, 0x99, 0x47, 0xd8, 0x50, 0x05, 0x1a, 0xb4, 0xb0, 0x3d, 0x32, 0xa7,
	0x86, 0xba, 0x85, 0x00, 0xa4, 0xf9, 0x99, 0x31, 0x74, 0xb0, 0xda, 0xa0, 0x81, 0x53, 0x6c, 0xdb,
	0xc3, 0x57, 0x58, 0x6d, 0xd2, 0x7f, 0xc6, 0x43, 0x73, 0x32, 0xb7, 0xb0, 0x2a, 0x51, 0x6a, 0x36,
	0x1e, 0x63, 0xcb, 0x56, 0x45, 0xd4, 0x86, 0x26, 0xb6, 0xac, 0x99, 0xa5, 0xca, 0x68, 0x1b, 0xda,
	0xc7, 0x78, 0x68, 0x39, 0x47, 0x78, 0xe8, 0xa8, 0xad, 0xfd, 0xbf, 0x3b, 0xd0, 0x18, 0xb9, 0x61,
	0xb8, 0x46, 0x65, 0x42, 0x8d, 0xca, 0xfa, 0xdc, 0x1d, 0x44, 0xe6, 0x0e, 0xfa, 0x33, 0x25, 0xd3,
	0x72, 0x85, 0x39, 0x1c, 0x42, 0x7b, 0x65, 0x0e, 0x7c, 0xd0, 0x5f, 0xae, 0xc7, 0x57, 0x72, 0x47,
	0xdf, 0x82, 0xe4, 0x7a, 0x1e, 0x59, 0x66, 0x7c, 0xd2, 0x77, 0xd7, 0x27, 0x0c, 0x19, 0x83, 0x06,
	0x20, 0xfb, 0xc4, 0x0b, 0x83, 0xa8, 0x7e, 0xd0, 0x19, 0x6e, 0x14, 0x10, 0x5d, 0xfb, 0x75, 0x10,
	0x86, 0x7c, 0xca, 0x6b, 0xd6, 0x7e, 0x12, 0x84, 0x21, 0xfa, 0x01, 0x5a, 0xe9, 0x55, 0x9e, 0xf9,
	0xf1, 0x5d, 0x54, 0x3b, 0xe2, 0xc5, 0xd2, 0x39, 0x85, 0x7e, 0x06, 0xc5, 0xf5, 0xae, 0xa3, 0xf8,
	0x2e, 0x24, 0xfe, 0x25, 0xe1, 0x83, 0xbe, 0x57, 0xb7, 0xfc, 0x15, 0x48, 0x4f, 0x29, 0xa1, 0x23,
	0xea, 0x05, 0x21, 0xd1, 0xda, 0x9b, 0x4e, 0xc9, 0x2a, 0x31, 0xba, 0xef, 0x52, 0x45, 0xb0, 0x69,
	0xdf, 0xe5, 0xec, 0x0c, 0xa8, 0x81, 0xfe, 0x91, 0x93, 0x34, 0xd3, 0x94, 0x4d, 0xbc, 0x55, 0x40,
	0xfa, 0x2f, 0xd0, 0xae, 0x5e, 0x49, 0xb7, 0x22, 0x8d, 0x68, 0x11, 0xf3, 0xf1, 0xfb, 0xf4, 0x99,
	0x38, 0xa2, 0x45, 0xac, 0xbf, 0x05, 0x89, 0x5f, 0xce, 0x1e, 0xb4, 0x4b, 0x2b, 0x28, 0x3d, 0xe6,
	0x89, 0x17, 0xa0, 0x6f, 0x00, 0xe2, 0x25, 0x49, 0xdc, 0x2c, 0x88, 0xa3, 0xd2, 0x61, 0x3e, 0xab,
	0x32, 0x83, 0x59, 0x19, 0x46, 0x2f, 0x41, 0x5e, 0x04, 0x61, 0xf6, 0xf0, 0x68, 0x94, 0xc5, 0xc6,
	0xc5, 0x57, 0xfd, 0x14, 0xe4, 0xf2, 0x9e, 0xdf, 0xa3, 0x75, 0xa5, 0x9c, 0xb8, 0xb6, 0xdc, 0x12,
	0x1a, 0x4c, 0x09, 0x2f, 0x40, 0xce, 0xdc, 0xb4, 0x62, 0xbd, 0xdb, 0x15, 0x9b, 0x32, 0x8d, 0x47,
	0xa6, 0x21, 0xd6, 0x99, 0x06, 0x55, 0xdd, 0xf9, 0x32, 0x0e, 0x03, 0xef, 0x9e, 0xaf, 0xbe, 0x34,
	0x0d, 0xda, 0xe3, 0x8c, 0x05, 0x74, 0x07, 0x5a, 0x2b, 0x35, 0x3d, 0x31, 0x1a, 0xa1, 0xce, 0xcb,
	0x1e, 0x77, 0x5f, 0xe3, 0x8a, 0xfa, 0xef, 0xa0, 0x54, 0xe5, 0xf6, 0x6e, 0x1b, 0xad, 0x6c, 0x58,
	0x5c, 0xb7, 0xe1, 0x0e, 0x34, 0xf2, 0x3c, 0xf0, 0xb9, 0x7d, 0xfe, 0x25, 0x40, 0xfb, 0x41, 0x98,
	0x3f, 0x42, 0x93, 0xe6, 0x96, 0x87, 0xfe, 0xe5, 0x3b, 0x84, 0xcc, 0x4a, 0xea, 0xc7, 0xd0, 0xa0,
	0x7f, 0xff, 0xff, 0x49, 0x7f, 0x88, 0xe7, 0x82, 0xbd, 0x75, 0x6c, 0x4a, 0x68, 0x0b, 0x3e, 0x55,
	0x4f, 0x45, 0xc6, 0x89, 0xfd, 0x7f, 0x84, 0x75, 0xee, 0xbe, 0x0d, 0xed, 0x95, 0xbb, 0xab, 0x02,
	0xea, 0x40, 0xcb, 0xc1, 0x43, 0xcb, 0xa0, 0x41, 0x91, 0xfa, 0xf6, 0x70, 0x34, 0xc2, 0x67, 0x8e,
	0xba, 0x45, 0xb3, 0x0c, 0x3c, 0x9a, 0x98, 0x53, 0x6a, 0xf5, 0x00, 0x92, 0x85, 0x5f, 0x9b, 0xaf,
	0xa9, 0xd3, 0xb7, 0xa0, 0x71, 0x62, 0x4e, 0x26, 0xaa, 0x44, 0x93, 0xed, 0xe3, 0xb9, 0xc3, 0x92,
	0x65, 0xb4, 0x03, 0xca, 0x70, 0x44, 0xdb, 0x4c, 0xb0, 0xf1, 0x0a, 0xab, 0x2d, 0xda, 0xca, 0xc2,
	0xa3, 0xd9, 0x74, 0x64, 0x4e, 0xb0, 0xda, 0xae, 0x3e, 0x17, 0x50, 0x3c, 0x2a, 0xbf, 0xce, 0xb1,
	0xed, 0xa8, 0x0a, 0xab, 0x33, 0x3f, 0x3b, 0xb3, 0xb0, 0x6d, 0xab, 0x9d, 0xa3, 0xce, 0x1b, 0x60,
	0x1b, 0x61, 0xbf, 0x0c, 0xff, 0x0b, 0x00, 0x00, 0xff, 0xff, 0xc2, 0xfb, 0x6b, 0x54, 0x49, 0x0a,
	0x00, 0x00,
}
