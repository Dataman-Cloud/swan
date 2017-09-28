package kvm

type Message struct {
	TaskId  string `json:"taskId"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (e *Executor) NewMessage(status, msg string) Message {
	return Message{
		TaskId:  e.taskId.GetValue(),
		Status:  status,
		Message: msg,
	}
}
