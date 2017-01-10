package config

const (
	API_PREFIX = "/v_beta"
)

type SwanMode string

const (
	Manager SwanMode = "manager"
	Agent   SwanMode = "agent"
	Mixed   SwanMode = "mixed"
)
