package event

const (
	EventTypeTaskAdd = "task_add"
	EventTypeTaskRm  = "task_rm"
)

type Event struct {
	Type    string
	Payload interface{}
}

func NewEvent(t string, payload interface{}) *Event {
	return &Event{
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
