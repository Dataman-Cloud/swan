package kvm

// Message is a message object which kvm executor send to mesos scheduler
type Message struct {
	TaskId  string `json:"taskId"`
	Status  string `json:"status"`
	Message string `json:"message"`
	VncAddr string `json:"vncAddr"`
}

func (e *Executor) NewMessage(status, msg, vnc string) Message {
	return Message{
		TaskId:  e.taskId.GetValue(),
		Status:  status,
		Message: msg,
		VncAddr: vnc,
	}
}

// FramwworkMessage is a message object which mesos scheduler send to kvm executor
type FrameworkMessage string

var (
	FMMsgShutDown = FrameworkMessage("SWAN_KVM_TASK_SHUTDOWN")
	FMMsgStartUp  = FrameworkMessage("SWAN_KVM_TASK_STARTUP")
	FMMsgSuspend  = FrameworkMessage("SWAN_KVM_TASK_SUSPEND")
	FMMsgResume   = FrameworkMessage("SWAN_KVM_TASK_RESUME")
)
