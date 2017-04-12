package event

import "fmt"

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

func (ue *UserEvent) String() string {
	return fmt.Sprintf("type: %s, param: %v", ue.Type, ue.Param)
}
