package types

type TaskInfoEvent struct {
	Ip        string
	TaskId    string
	AppId     string
	Port      string
	State     string
	Healthy   bool
	ClusterId string
	RunAs     string
}

type AppInfoEvent struct {
	AppId     string
	Name      string
	State     string
	ClusterId string
	RunAs     string
}
