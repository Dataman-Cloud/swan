package types

// HealthCheck is the definition for an application health check
type HealthCheck struct {
	ID                     string   `json:"id"`
	Address                string   `json:"url"`
	TaskID                 string   `json:"task_id"`
	AppID                  string   `json:"app_id"`
	Protocol               string   `json:"protocol,omitempty"`
	Port                   *int     `json:"port,omitempty"`
	PortIndex              *int     `json:"portIndex,omitempty"`
	Command                *Command `json:"command,omitempty"`
	Path                   *string  `json:"path,omitempty"`
	MaxConsecutiveFailures *int     `json:"maxConsecutiveFailures,omitempty"`
	GracePeriodSeconds     int64    `json:"gracePeriodSeconds,omitempty"`
	IntervalSeconds        int64    `json:"intervalSeconds,omitempty"`
	TimeoutSeconds         int64    `json:"timeoutSeconds,omitempty"`
}

type Command struct {
	Value string `json:"value"`
}
