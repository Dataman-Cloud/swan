package types

type TaskInfoEvent struct {
	IP         string  `json:"ip"`
	TaskID     string  `json:"taskID"`
	AppID      string  `json:"appID"`
	AppVersion string  `json:"appVersion"`
	VersionID  string  `json:"versionID"`
	Port       uint32  `json:"port"`
	PortName   string  `json:"portName"`
	State      string  `json:"state"`
	Healthy    bool    `json:"healthy"`
	ClusterID  string  `json:"clusterID"`
	RunAs      string  `json:"runAs"`
	Mode       string  `json:"mode"`
	Weight     float64 `json:"weight"`
	AppName    string  `json:"appName"`
	SlotIndex  int     `json:"slotIndex"`
}

type AppInfoEvent struct {
	AppID     string `json:"appID"`
	Name      string `json:"name"`
	State     string `json:"state"`
	ClusterID string `json:"clusterID"`
	RunAs     string `json:"runAs"`
	Mode      string `json:"mode"`
}
