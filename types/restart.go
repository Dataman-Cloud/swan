package types

type RestartPolicy struct {
	Attempts int     `json:"attempts"`
	Delay    float64 `json:"delay"`
}
