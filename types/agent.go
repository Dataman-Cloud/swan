package types

import (
	"errors"
	"time"
)

// Agent is a db swan Agent
type Agent struct {
	ID         string    `json:"id"`
	SysInfo    *SysInfo  `json:"sysinfo"`
	CreatedAt  time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
}

func (a *Agent) Valid() error {
	if a.ID == "" {
		return errors.New("agent ID required")
	}
	if a.SysInfo == nil {
		return errors.New("agent sysinfo required")
	}
	return nil
}
