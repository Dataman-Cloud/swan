package agent

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HTTPServer instance
type HTTPServer struct {
	listener string
	agentRef *Agent
	engine   *gin.Engine
}

// NewHTTPServer is singleton instance func
func NewHTTPServer(listener string, a *Agent) *HTTPServer {
	aas := &HTTPServer{
		listener: listener,
		agentRef: a,
	}
	aas.engine = gin.Default()

	aas.engine.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	aas.engine.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	aas.engine.GET("/agents", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"agents": aas.agentRef.SerfServer.SerfNode.Members()})
	})

	proxyRouter := aas.engine.Group("/proxy")
	a.Janitor.ApiServe(proxyRouter)

	aas.engine.GET("/dns", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"records": aas.agentRef.Resolver.AllRecords()})
	})

	return aas
}

// Start func
func (aas *HTTPServer) Start() error {
	return aas.engine.Run(aas.listener)
}
