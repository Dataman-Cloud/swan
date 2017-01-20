package types

type ServiceDiscovery struct {
	TaskID           string             `json:"taskID"`
	AppID            string             `json:"appID"`
	AppMode          string             `json:"appMode"`
	IP               string             `json:"ip"`
	TaskPortMappings []*TaskPortMapping `json:"taskPortMappings"`
	URL              string             `json:"url"`
}

type TaskPortMapping struct {
	HostPort      int32  `json:"hostPort"`
	ContainerPort int32  `json:"containerPort"`
	Name          string `json:"name"`
	Protocol      string `json:"protocol"`
}
