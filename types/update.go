package types

type UpdatePolicy struct {
	UpdateDelay int `json:"updateDelay"`
	MaxRetries  int `json:"maxRetries"`
}
