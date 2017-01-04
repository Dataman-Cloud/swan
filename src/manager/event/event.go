package event

import (
	"github.com/satori/go.uuid"
)

const (
	EventTypeTaskAdd = "task_add"
	EventTypeTaskRm  = "task_rm"
	EventTypeAppAdd  = "app_add"
	EventTypeAppRm   = "app_rm"
)

type Event struct {
	Id      string
	Type    string
	Payload interface{}
}

func NewEvent(t string, payload interface{}) *Event {
	return &Event{
		Id:      uuid.NewV4().String(),
		Type:    t,
		Payload: payload,
	}
}

type TaskInfo struct {
	Ip     string
	TaskId string
	Port   string
	Type   string // a or srv
}
