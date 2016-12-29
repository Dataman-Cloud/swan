package event

const (
	UserEventType1 = "1"
	UserEventType2 = "2"
)

type UserEvent struct {
	Type  string
	Param interface{}
}

func (ue *UserEvent) GetEventType() string {
	return ue.Type
}

func (ue *UserEvent) GetEvent() interface{} {
	return ue.Param
}
