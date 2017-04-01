package agent

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

type HttpServer struct {
	listener string
	agentRef *Agent
	engine   *gin.Engine
}

func NewHttpServer(listener string, a *Agent) *HttpServer {
	aas := &HttpServer{
		listener: listener,
		agentRef: a,
	}
	aas.engine = gin.Default()

	aas.engine.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	aas.engine.GET("/agents", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"agents": aas.agentRef.SerfServer.SerfNode.Members()})
	})

	return aas
}

func (aas *HttpServer) Start(ctx context.Context, started chan bool) error {
	errCh := make(chan error)
	go func() {
		errCh <- aas.engine.Run(aas.listener)
	}()

	go func() {
		started <- true
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
