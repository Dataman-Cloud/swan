package types

type Check struct {
	ID          string   `json:"id"`
	Protocol    string   `json:"protocol"`
	Address     string   `json:"address"`
	Port        int      `json:"port"`
	Command     *Command `json:"command"`
	Path        string   `json:"path"`
	MaxFailures int      `json:"max_failures"`
	Interval    int      `json:"interval"`
	Timeout     int      `json:"timeout"`
	TaskID      string   `json:"task_id"`
	AppID       string   `json:"app_id"`
}
