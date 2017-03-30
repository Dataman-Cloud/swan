package agent

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AgentApiServer struct {
	listener string
	agentRef *Agent
	engine   *gin.Engine
}

func NewAgentApiServer(listener string, a *Agent) *AgentApiServer {
	aas := &AgentApiServer{
		listener: listener,
		agentRef: a,
	}
	aas.engine = gin.Default()

	aas.engine.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return aas
}

func (aas *AgentApiServer) Start() error {
	return aas.engine.Run(aas.listener)
}
