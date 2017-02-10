package config

type SwanMode string

const (
	Manager SwanMode = "manager"
	Agent   SwanMode = "agent"
	Mixed   SwanMode = "mixed"
)
