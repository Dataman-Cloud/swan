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

func (m *MesosAgent) Match(filter map[string][]string) bool {
	var n int
	for k, vals := range filter {
		if v, ok := m.Attrs[k]; ok {
			if in(v, vals) {
				n++
			}
		}
	}

	return n == len(filter)
}

func in(v string, vals []string) bool {
	for i := 0; i < len(vals); i++ {
		if v == vals[i] {
			return true
		}
	}

	return false
}
