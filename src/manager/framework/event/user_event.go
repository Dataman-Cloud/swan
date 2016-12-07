package event

import ()

type UserEventAction string

const (
	UserEventActionType1 UserEventAction = "1"
	UserEventActionType2 UserEventAction = "2"
)

type UserEvent struct {
	AppId  string
	RunAs  string
	Action UserEventAction
	Param  interface{}
}
