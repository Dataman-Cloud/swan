package types

type TaskInfoEvent struct {
	IP        string `json:"ip"`
	TaskID    string `json:"taskID"`
	AppID     string `json:"appID"`
	Port      string `json:"port"`
	State     string `json:"state"`
	Healthy   bool   `json:"healthy"`
	ClusterID string `json:"clusterID"`
	RunAs     string `json:"runAs"`
}

type AppInfoEvent struct {
	AppID     string `json:"appID"`
	Name      string `json:"name"`
	State     string `json:"state"`
	ClusterID string `json:"clusterID"`
	RunAs     string `json:"runAs"`
}
