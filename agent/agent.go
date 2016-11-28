package agent

import (
	"github.com/Dataman-Cloud/swan/util"
)

type Agent struct {
	Config util.SwanConfig
}

func New(config util.SwanConfig) (*Agent, error) {
	agent := &Agent{
		Config: config,
	}

	return agent, nil
}

func (agent *Agent) Start() error {
	return nil
}
