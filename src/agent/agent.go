package agent

import (
	"github.com/Dataman-Cloud/swan/src/config"
)

type Agent struct {
	Config config.SwanConfig
}

func New(config config.SwanConfig) (*Agent, error) {
	agent := &Agent{
		Config: config,
	}

	return agent, nil
}

func (agent *Agent) Start() error {
	return nil
}
