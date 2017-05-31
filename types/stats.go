package types

type Stats struct {
	ClusterID string `json:"clusterID"`

	AppCount  int `json:"appCount"`
	TaskCount int `json:"taskCount"`

	Created float64 `json:"created"`

	Master string `json:"master"`
	Slaves string `json:"slaves"`

	Attributes []map[string]interface{} `json:"attributes"`

	TotalCpu  float64 `json:"totalCpu"`
	TotalMem  float64 `json:"totalMem"`
	TotalDisk float64 `json:"totalDisk"`

	CpuTotalOffered  float64 `json:"cpuTotalOffered"`
	MemTotalOffered  float64 `json:"memTotalOffered"`
	DiskTotalOffered float64 `json:"diskTotalOffered"`

	CpuTotalUsed  float64 `json:"cpuTotalUsed"`
	MemTotalUsed  float64 `json:"memTotalUsed"`
	DiskTotalUsed float64 `json:"diskTotalUsed"`

	AppStats map[string]int `json:"appStats,omitempty"`
}
