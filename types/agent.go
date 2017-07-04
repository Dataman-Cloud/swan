package types

import (
	"errors"
	"time"
)

// Agent is a db swan Agent
type Agent struct {
	ID        string    `json:"id"`
	Hostname  string    `json:"hostname"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *Agent) Valid() error {
	if a.ID == "" {
		return errors.New("agent ID required")
	}
	return nil
}
