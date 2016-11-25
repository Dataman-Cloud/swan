package agent

import (
	"github.com/Dataman-Cloud/swan/util"
)

type Agent struct {
	Config util.SwanConfig
}

func New(config util.SwanConfig) *Agent {
	agent := &Agent{
		Config: config,
	}

	return agent
}

func (agent *Agent) Start() {}
func (agent *Agent) Run() error {
	return nil
}
