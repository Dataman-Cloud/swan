package types

type MesosAgent struct {
	ID    string            `json:"id"`
	IP    string            `json:"ip"`
	CPUs  float64           `json:"cpus"`
	Mem   float64           `json:"mem"`
	GPUs  float64           `json:"gpus"`
	Disk  float64           `json:"disk"`
	Ports int64             `json:"ports"`
	Attrs map[string]string `json:"attrs"`
}
