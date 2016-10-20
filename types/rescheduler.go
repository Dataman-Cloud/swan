package types

type ReschedulerMsg struct {
	AppID  string
	TaskID string
	Err    chan error
}
