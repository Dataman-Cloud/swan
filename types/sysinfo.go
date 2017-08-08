package types

type SysInfo struct {
	Hostname   string              `json:"hostname"`
	OS         string              `json:"os"`
	Uptime     string              `json:"uptime"`
	UnixTime   int64               `json:"unixtime"`
	LoadAvg    float64             `json:"loadavg"`
	CPU        CPUInfo             `json:"cpu"`
	Memory     MemoryInfo          `json:"memory"`
	Containers ContainersInfo      `json:"containers"`
	IPs        map[string][]string `json:"ips"` // inet name -> ips
	Listenings []int64             `json:"listenings"`
}

type MemoryInfo struct {
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
}

type CPUInfo struct {
	Processor int64   `json:"processor"`
	Physical  int64   `json:"physical"`
	Used      float64 `json:"used"` // 1 - idle
}

type ContainersInfo struct {
	Total   int64 `json:"total"`
	Running int   `json:"running"`
	Stopped int   `json:"stopped"`
	Killed  int   `json:"killed"`
	Paused  int   `json:"paused"`
}
